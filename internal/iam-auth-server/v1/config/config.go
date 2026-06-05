package config

import (
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/analytics"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/load/cache"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"time"
)

type Config struct {
	Server    ServerConfig                 `mapstructure:"server"`
	Grpc      GrpcConfig                   `mapstructure:"grpc"`
	Log       *log.Options                 `mapstructure:"log"`
	Redis     *genericoptions.RedisOptions `mapstructure:"redis"`
	Analytics *analytics.AnalyticsOptions  `mapstructure:"analytics"`
	Cache     *cache.Options               `mapstructure:"cache"`
}

type ServerConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	HealthPath        string        `mapstructure:"health_path"`
	HealthAddr        string        `mapstructure:"health_addr"`
	Mode              string        `mapstructure:"mode"` // debug/release/test
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	GraceTimeout      time.Duration `mapstructure:"grace_timeout"`
	EnableRedisSink   bool          `mapstructure:"enable_redis_sink"`
	RedisLogKeyPrefix string        `mapstructure:"redis_key_prefix"`
	SinkLevel         string        `mapstructure:"sink_level"`
	LoadCacheTTL      time.Duration `mapstructure:"load_cache_ttl"`
}
type GrpcConfig struct {
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	EtcdServerDiscovery bool   `mapstructure:"etcd_server_discovery"`
	ServiceName         string `mapstructure:"service_name"`
}

func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:              "0.0.0.0",
			Port:              8080,
			Mode:              "debug",
			ReadTimeout:       time.Second * 15,
			WriteTimeout:      time.Second * 15,
			IdleTimeout:       time.Second * 60,
			GraceTimeout:      time.Second * 60,
			EnableRedisSink:   false,
			RedisLogKeyPrefix: "iam-auth",
			SinkLevel:         "info",
		},
		Grpc: GrpcConfig{
			Host:                "0.0.0.0",
			Port:                8080,
			EtcdServerDiscovery: false,
			ServiceName:         "",
		},
		Log:       log.NewOptions(),
		Redis:     genericoptions.NewRedisOptions(),
		Analytics: analytics.NewAnalyticsOptions(),
	}
}
