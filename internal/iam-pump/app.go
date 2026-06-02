package iampump

import (
	"github.com/sis-shen/sup-iam/internal/iam-pump/options"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/sis-shen/sup-iam/internal/pkg/server"
)

type App struct {
	name     string
	basename string
	opts     *options.Options
}

func NewApp(basename string) *App {
	opts := options.Load("")
	return &App{
		name:     "Iam Pump Server",
		basename: basename,
		opts:     opts,
	}
}

func (app *App) Start() error {
	log.Init(app.opts.Log)
	defer log.Flush()

	stopChan := server.SetupSignalHandler()

	return Run(app.opts, stopChan)
}
