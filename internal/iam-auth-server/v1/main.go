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

func main() {
	var showVersion bool
	pflag.BoolVarP(&showVersion, "version", "v", false, "show version")
	pflag.Parse()
	if showVersion {
		//显示版本信息后退出
		fmt.Printf("Version: %s\nBuildTime: %s\nCommitHash: %s\n", Version, BuildTime, CommitHash)
		return
	}

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
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", conf.Grpc.Host, conf.Grpc.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatalf("Fail to connect grpc server: %v", err.Error())
	}

	defer conn.Close()
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
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Errorf("fail to shutdown http server: %v", err)
		}

		return nil
	})

	// 8.等待所有goroutine结束
	if err := g.Wait(); err != nil {
		logger.Fatalf("fail to wait: %v", err)
	}

	logger.Infof("finish graceful shutdown")
}
