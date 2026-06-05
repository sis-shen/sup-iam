package iampump

import (
	"github.com/sis-shen/sup-iam/internal/iam-pump/options"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
)

import genericapiserver "github.com/sis-shen/sup-iam/internal/pkg/server"

func Run(o *options.Options, stopChan <-chan struct{}) error {
	go genericapiserver.ServeHealthCheck(o.HealthCheckPath, o.HealthCheckAddress)

	server, err := createPumpServer(o)
	if err != nil {
		log.Errorf("Error starting pump server: %v", err)
		return nil
	}
	return server.PrepareRun().Run(stopChan)
}
