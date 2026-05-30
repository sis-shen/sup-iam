package main

import (
	"context"
	"fmt"
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
