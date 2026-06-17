package server

import (
	"net/http"
	"time"

	"github.com/sis-shen/sup-iam/internal/pkg/log"
)

func ServeHealthCheck(path string, address string) {
	// 使用自定义 ServeMux，而不是默认的 DefaultServeMux
	mux := http.NewServeMux()
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"healthy": true}`))
	})

	// 创建自定义 Server，配置更好的参数
	srv := &http.Server{
		Addr:           address,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
		// 优化：禁用 keep-alive 或设置更长的超时
	}

	log.Infof("Starting health check server on %s", address)
	if err := srv.ListenAndServe(); err != nil {
		log.Errorf("Error starting health check: %v", err)
	}
}
