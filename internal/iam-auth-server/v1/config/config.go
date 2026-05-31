package config

import (
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"time"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Grpc   GrpcConfig   `mapstructure:"grpc"`
	Log    *log.Options `mapstructure:"log"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"` // debug/release/test
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	GraceTimeout    time.Duration `mapstructure:"grace_timeout"`
	EnableRedisSink bool          `mapstructure:"enable_redis_sink"`
	RedisKeyPrefix  string        `mapstructure:"redis_key_prefix"`
	SinkLevel       string        `mapstructure:"sink_level"`
}
type GrpcConfig struct {
	Host                string `mapstructure:"host"`
	Port                int    `mapstructure:"port"`
	EtcdServerDiscovery bool   `mapstructure:"etcd_server_discovery"`
	ServiceName         string `mapstructure:"service_name"`
}

type RedisConfig struct {
	Host                string        `mapstructure:"host"`
	Port                int           `mapstructure:"port"`
	Password            string        `mapstructure:"password"`
	DatabaseName        string        `mapstructure:"database_name"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
	//连接配置
	PoolSize        int           `mapstructure:"pool_size"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	// 超时配置
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
}
