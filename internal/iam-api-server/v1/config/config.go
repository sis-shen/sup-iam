package config

import (
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"time"
)

type AppConfig struct {
	Server     *ServerConfig                `mapstructure:"server"`
	JWT        *JWTConfig                   `mapstructure:"jwt"`
	MySQL      *MySQLConfig                 `mapstructure:"mysql"`
	Redis      *genericoptions.RedisOptions `mapstructure:"redis"`
	Log        *log.Options                 `mapstructure:"log"`
	GrpcConfig *GrpcConfig                  `mapstructure:"grpc"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host              string        `mapstructure:"host"`
	Port              int           `mapstructure:"port"`
	HealthPath        string        `mapstructure:"health_path"`
	HealthAddr        string        `mapstructure:"health_addr"`
	Mode              string        `mapstructure:"mode"` // debug/release/test
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	BlackListTTL      time.Duration `mapstructure:"black_list_ttl"`
	GraceTimeout      time.Duration `mapstructure:"grace_timeout"`
	EnableRedisSink   bool          `mapstructure:"enable_redis_sink"`
	RedisLogKeyPrefix string        `mapstructure:"redis_key_prefix"`
	SinkLevel         string        `mapstructure:"sink_level"`
}

type JWTConfig struct {
	SecretKey              string        `mapstructure:"secret_key"`
	AccessTokenExpireTime  time.Duration `mapstructure:"access_token_expire_time"`
	RefreshTokenExpireTime time.Duration `mapstructure:"refresh_token_expire_time"`
	UserIDKey              string        `mapstructure:"user_id_key"`
	TokenLookup            string        `mapstructure:"token_lookup"`
	Issuer                 string        `mapstructure:"issuer"`
	SkipPaths              []string      `mapstructure:"skip_paths"`
}

type MySQLConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	DatabaseName    string        `mapstructure:"database_name"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	MaxRetries      int           `mapstructure:"max_retries"`
}

type GrpcConfig struct {
	Host                string        `mapstructure:"host"`
	Port                int           `mapstructure:"port"`
	EtcdServerDiscovery bool          `mapstructure:"etcd_server_discovery"`
	EtcdHost            string        `mapstructure:"etcd_host"`
	EtcdPort            int           `mapstructure:"etcd_port"`
	ServiceName         string        `mapstructure:"service_name"`
	LeaseTTL            time.Duration `mapstructure:"lease_ttl"`
	ServiceAddress      string        `mapstructure:"service_address"`
}
