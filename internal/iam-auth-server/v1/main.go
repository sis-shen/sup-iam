package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/config"
	server "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/go"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/rpc"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/service"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v1"
	"github.com/spf13/pflag"
	etcdclient "go.etcd.io/etcd/client/v3"
	etcdresolver "go.etcd.io/etcd/client/v3/naming/resolver"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Version    string
	BuildTime  string
	CommitHash string
)

type ShutdownResources struct {
	httpServer *http.Server
	grpcConn   *grpc.ClientConn
	etcdClient *etcdclient.Client
}

var shutdownRes ShutdownResources

func main() {
	var showVersion bool
	pflag.BoolVarP(&showVersion, "version", "v", false, "show version")
	pflag.Parse()
	if showVersion {
		//显示版本信息后退出
		fmt.Printf("Version: %s\nBuildTime: %s\nCommitHash: %s\n", Version, BuildTime, CommitHash)
		return
	}

	shutdownRes = ShutdownResources{}
	//======== 1.加载配置
	conf, err := config.Load("")
	if err != nil {
		fmt.Printf("配置文件加载失败: %v", err.Error())
		return
	}

	errs := conf.Log.Validate()
	if errs != nil && len(errs) > 0 {
		fmt.Printf("fail to validate log: %v", errs)
		return
	}

	//====== 2. 加载组件
	logger := log.New(conf.Log)

	if conf.Server.EnableRedisSink {
		redisCli, err := initRedisWithHealthCheck(conf.Redis)
		if err != nil {
			logger.Fatalf("fail to init redis client: %v", err)
			return
		}
		var level log.Level
		if err := level.UnmarshalText([]byte(conf.Server.SinkLevel)); err != nil {
			level = log.InfoLevel
		}
		logger = log.WrapWithRedis(logger, redisCli,
			conf.Server.RedisKeyPrefix,
			level)
	}
	conn, err := newGrpcConn(conf.Grpc, logger)
	if err != nil {
		logger.Errorf("fail to create grpc conn: %v", err)
		return
	}
	shutdownRes.grpcConn = conn

	client := pbv1.NewAuthQueryServiceClient(conn)

	// =======3. 创建errgroup管理
	g, ctx := errgroup.WithContext(context.Background())

	// ========= 4.初始化router
	rpcClient := rpc.NewGRpcClient(client)
	authCase := service.NewAuthCase(rpcClient)
	authAPI := server.NewAuthVerifyAPI(authCase, logger)
	routes := server.ApiHandleFunctions{
		AuthVerifyAPI: *authAPI,
	}
	router := server.NewRouter(routes)

	//======== 5.启动HTTPServer
	address := fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)
	httpServer := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  conf.Server.ReadTimeout,
		WriteTimeout: conf.Server.WriteTimeout,
		IdleTimeout:  conf.Server.IdleTimeout,
	}

	g.Go(func() error {
		logger.Infof("starting http server on %s", address)
		shutdownRes.httpServer = httpServer
		if err := httpServer.ListenAndServe(); err != nil {
			return fmt.Errorf("fail to start http server: %v", err)
		}
		return nil
	})

	// ====== 6.监听系统退出信号
	g.Go(func() error {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		select {
		case sig := <-quit:
			logger.Infof("receive signal: %v, start to graceful shutdown", sig)
		case <-ctx.Done():
			logger.Infof("unknown error in goroutine")
		}

		//关停server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), conf.Server.GraceTimeout)
		defer shutdownCancel()

		if shutdownRes.httpServer != nil {
			if err := shutdownRes.httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Errorf("fail to shutdown http server: %v", err)
			} else {
				logger.Infof("http server shutdown successfully")
			}
		}

		// 2. 关闭 gRPC 客户端连接
		if shutdownRes.grpcConn != nil {
			if err := shutdownRes.grpcConn.Close(); err != nil {
				logger.Errorf("fail to close grpc connection: %v", err)
			} else {
				logger.Infof("grpc connection closed")
			}
		}

		// 3. 关闭 etcd 客户端
		if shutdownRes.etcdClient != nil {
			if err := shutdownRes.etcdClient.Close(); err != nil {
				logger.Errorf("fail to close etcd client: %v", err)
			} else {
				logger.Infof("etcd client closed")
			}
		}

		return nil
	})

	// 8.等待所有goroutine结束
	if err := g.Wait(); err != nil {
		logger.Fatalf("fail to wait: %v", err)
	}

	logger.Infof("finish graceful shutdown")
}

func newGrpcConn(conf config.GrpcConfig, logger log.Logger) (*grpc.ClientConn, error) {
	if conf.EtcdServerDiscovery {
		//连接到etcd
		etcdCli, err := etcdclient.NewFromURL(fmt.Sprintf("%s:%d", conf.Host, conf.Port))
		if err != nil {
			logger.Fatalf("fail to create etcd client: %v", err)
			return nil, err
		}

		// 创建etcd resolver
		builder, err := etcdresolver.NewBuilder(etcdCli)
		if err != nil {
			logger.Fatalf("fail to create etcd resolver: %v", err)
			return nil, err
		}
		shutdownRes.etcdClient = etcdCli

		target := fmt.Sprintf("etcd:///%s", conf.ServiceName)
		conn, err := grpc.NewClient(
			target,
			grpc.WithResolvers(builder),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		)

		if err != nil {
			logger.Fatalf("fail to create grpc client: %v", err)
			return nil, err
		}

		return conn, nil
	} else {
		conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", conf.Host, conf.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatalf("Fail to connect grpc server: %v", err.Error())
			return nil, err
		}

		return conn, nil
	}
}

// initRedisWithHealthCheck Redis 连接（带健康检查）
func initRedisWithHealthCheck(conf config.RedisConfig) (*redis.Client, error) {
	client, err := initRedis(conf)
	if err != nil {
		return nil, err
	}

	// 启动健康检查 goroutine
	go func() {
		ticker := time.NewTicker(conf.HealthCheckInterval)
		defer ticker.Stop()

		ctx := context.Background()
		for range ticker.C {
			if err := client.Ping(ctx).Err(); err != nil {
				fmt.Printf("Redis health check failed: %v\n", err)
			}
		}
	}()

	return client, nil
}

func initRedis(conf config.RedisConfig) (*redis.Client, error) {
	// 构建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		Password: conf.Password,
		DB:       getRedisDB(conf.DatabaseName), // 将数据库名称转换为数字

		// 连接池配置
		PoolSize:        conf.PoolSize,        // 连接池大小
		MinIdleConns:    conf.MinIdleConns,    // 最小空闲连接数
		MaxIdleConns:    conf.MaxIdleConns,    // 最大空闲连接数
		ConnMaxIdleTime: conf.ConnMaxIdleTime, // 连接最大空闲时间
		ConnMaxLifetime: conf.ConnMaxLifetime, // 连接最大生命周期

		// 超时配置
		DialTimeout:  conf.DialTimeout,  // 连接超时
		ReadTimeout:  conf.ReadTimeout,  // 读超时
		WriteTimeout: conf.WriteTimeout, // 写超时
		PoolTimeout:  conf.PoolTimeout,  // 连接池获取连接超时
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

// getRedisDB 将数据库名称转换为 Redis DB 编号
// 支持字符串格式: "0", "1", "default" 等
func getRedisDB(dbName string) int {
	if dbName == "" {
		return 0
	}

	// 尝试转换为整数
	var db int
	_, err := fmt.Sscanf(dbName, "%d", &db)
	if err != nil {
		// 如果转换失败，使用默认值 0
		return 0
	}

	return db
}
