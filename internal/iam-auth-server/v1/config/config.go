package config

import (
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"time"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Grpc   GrpcConfig   `mapstructure:"grpc"`
	Log    *log.Options `mapstructure:"log"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug/release/test
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	GraceTimeout time.Duration `mapstructure:"grace_timeout"`
}
type GrpcConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}
