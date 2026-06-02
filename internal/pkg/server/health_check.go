package server

import (
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"net/http"
)

func ServeHealthCheck(path string, address string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"healthy": true}`))
	})

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Errorf("Error starting health check: %v", err)
	}
}
