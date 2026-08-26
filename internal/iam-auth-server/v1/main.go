package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"google.golang.org/grpc/backoff"
	"time"

	_ "net/http/pprof"

	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/analytics"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/config"
	server "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/go"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/load"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/load/cache"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/rpc"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/service"
	storeredis "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/storage/redis"
	"github.com/sis-shen/sup-iam/internal/pkg/keys"
	genericapiserver "github.com/sis-shen/sup-iam/internal/pkg/server"
	pkgtls "github.com/sis-shen/sup-iam/internal/pkg/tls"
	"google.golang.org/grpc/encoding/gzip"

	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v2"
	"github.com/spf13/pflag"
	etcdclient "go.etcd.io/etcd/client/v3"
	etcdresolver "go.etcd.io/etcd/client/v3/naming/resolver"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Version    string
	BuildTime  string
	CommitHash string
)

type ShutdownResources struct {
	httpServer       *http.Server
	httpsServer      *http.Server
	grpcConn         *grpc.ClientConn
	etcdClient       *etcdclient.Client
	analyticsManager *analytics.Analytics
	loader           *load.LoadManager
	ctx              context.Context
}

var shutdownRes ShutdownResources

const startDelay time.Duration = time.Minute * 2

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
	shutdownRes.ctx = context.Background()
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

	if conf.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if conf.Server.Mode == "test" {
		gin.SetMode(gin.TestMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	go func() error {
		return http.ListenAndServe("0.0.0.0:6060", nil) // 独立端口
	}()

	//====== 2. 加载组件
	logger := log.New(conf.Log)
	redisCluster, err := storeredis.Init(conf.Redis)
	if err != nil {
		logger.Errorf("fail to init redis cluster: %v", err)
		return
	}
	err = redisCluster.EnsureConnection()

	if err != nil {
		logger.Errorf("fail to connect to redis cluster: %v", err)
		return
	}
	analyticsManager := analytics.NewAnalytics(conf.Analytics, redisCluster)
	err = analyticsManager.Start()
	if err != nil {
		logger.Errorf("fail to start analytics manager: %v", err)
		return
	}

	shutdownRes.analyticsManager = analyticsManager

	if conf.Server.EnableRedisSink {
		redisCli := redisCluster.DB()

		var level log.Level
		if err := level.UnmarshalText([]byte(conf.Server.SinkLevel)); err != nil {
			level = log.InfoLevel
		}
		logger = log.WrapWithRedis(logger, redisCli,
			conf.Server.RedisLogKeyPrefix,
			level)
	}

	logger.Infof("开始休眠，等待api server 启动,预计睡眠: %f 分钟", startDelay.Minutes())
	time.Sleep(startDelay)
	logger.Infof("停止休眠")
	conn, err := newGrpcConn(conf.Grpc, logger)
	if err != nil {
		logger.Errorf("fail to create grpc conn: %v", err)
		return
	}
	shutdownRes.grpcConn = conn

	client := pbv2.NewAuthQueryServiceClient(conn)
	rpcClient := rpc.NewGRpcClient(client)

	cacheIns, err := cache.InitSingleton(rpcClient, conf.Cache)

	loader := load.NewLoadManager(shutdownRes.ctx, cacheIns, conf.Server.LoadCacheTTL)
	err = loader.Start()
	if err != nil {
		logger.Errorf("fail to start load manager: %v", err)
	}

	shutdownRes.loader = loader

	keysIns := &keys.Keys{}

	// =======3. 创建errgroup管理
	g, ctx := errgroup.WithContext(context.Background())

	// ========= 4.初始化router
	authCase := service.NewAuthCase(cacheIns, keysIns, analyticsManager)
	authAPI := server.NewAuthVerifyAPI(authCase, logger)
	routes := server.ApiHandleFunctions{
		AuthVerifyAPI: *authAPI,
	}
	router := server.NewRouter(routes)

	//======== 5.启动HTTPServer
	go genericapiserver.ServeHealthCheck(conf.Server.HealthPath, conf.Server.HealthAddr)

	// 预加载 TLS 配置，失败则直接中止启动（与 redis/grpc 等组件启动失败的处理方式一致）
	var tlsConfig *tls.Config
	if conf.Server.TLS.Enabled {
		tlsConfig, err = pkgtls.NewServerTLSConfig(conf.Server.TLS.CertFile, conf.Server.TLS.KeyFile)
		if err != nil {
			logger.Errorf("fail to create tls config: %v", err)
			return
		}
	}

	address := fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)
	httpServer := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  conf.Server.ReadTimeout,
		WriteTimeout: conf.Server.WriteTimeout,
		IdleTimeout:  conf.Server.IdleTimeout,
	}
	shutdownRes.httpServer = httpServer

	g.Go(func() error {
		logger.Infof("starting http server on %s", address)
		if err := httpServer.ListenAndServe(); err != nil {
			return fmt.Errorf("fail to start http server: %v", err)
		}
		return nil
	})

	// 当 server.tls.enabled=true 时，8443 HTTPS 与 8080 HTTP 同时监听
	if conf.Server.TLS.Enabled {
		tlsAddress := fmt.Sprintf("%s:%d", conf.Server.Host, 8443)
		httpsServer := &http.Server{
			Addr:         tlsAddress,
			Handler:      router,
			TLSConfig:    tlsConfig,
			ReadTimeout:  conf.Server.ReadTimeout,
			WriteTimeout: conf.Server.WriteTimeout,
			IdleTimeout:  conf.Server.IdleTimeout,
		}
		shutdownRes.httpsServer = httpsServer

		g.Go(func() error {
			logger.Infof("starting https server on %s", tlsAddress)
			if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
				return fmt.Errorf("fail to start https server: %v", err)
			}
			return nil
		})
	}

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

		if shutdownRes.ctx != nil {
			shutdownRes.ctx.Done()
		}

		if shutdownRes.loader != nil {
			shutdownRes.loader.ShutDown()
		}

		// 关闭其它组件
		if shutdownRes.analyticsManager != nil {
			analyticsManager.Stop()
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

		if shutdownRes.httpsServer != nil {
			if err := shutdownRes.httpsServer.Shutdown(shutdownCtx); err != nil {
				logger.Errorf("fail to shutdown https server: %v", err)
			} else {
				logger.Infof("https server shutdown successfully")
			}
		}

		// 关闭 gRPC 客户端连接
		if shutdownRes.grpcConn != nil {
			if err := shutdownRes.grpcConn.Close(); err != nil {
				logger.Errorf("fail to close grpc connection: %v", err)
			} else {
				logger.Infof("grpc connection closed")
			}
		}

		//  关闭 etcd 客户端
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
			grpc.WithDefaultCallOptions(
				grpc.UseCompressor(gzip.Name),
				grpc.MaxCallRecvMsgSize(conf.MaxCallRecvMsgSize)),
		)

		if err != nil {
			logger.Fatalf("fail to create grpc client: %v", err)
			return nil, err
		}

		return conn, nil
	} else {
		// 配置重试选项
		retryPolicy := `{
        "methodConfig": [{
            "name": [{"service": ""}],
            "retryPolicy": {
                "MaxAttempts": 5,
                "InitialBackoff": "1s",
                "MaxBackoff": "30s",
                "BackoffMultiplier": 2.0,
                "RetryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED", "INTERNAL"]
            }
        }]
    }`

		conn, err := grpc.NewClient(
			fmt.Sprintf("%s:%d", conf.Host, conf.Port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(retryPolicy),
			grpc.WithDefaultCallOptions(
				grpc.UseCompressor(gzip.Name),
				grpc.MaxCallRecvMsgSize(conf.MaxCallRecvMsgSize),
			),
			// 连接参数配置 - 控制连接建立的重试
			grpc.WithConnectParams(grpc.ConnectParams{
				Backoff: backoff.Config{
					BaseDelay:  1 * time.Second,
					Multiplier: 1.6,
					Jitter:     0.2,
					MaxDelay:   120 * time.Second,
				},
				MinConnectTimeout: 5 * time.Second,
			}),
		)
		if err != nil {
			logger.Fatalf("Fail to connect grpc server: %v", err.Error())
			return nil, err
		}

		return conn, nil
	}
}
