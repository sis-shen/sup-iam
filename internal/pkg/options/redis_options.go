package options

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type RedisOptions struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`

	Addrs []string `mapstructure:"addrs"`

	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`

	DB                  int           `mapstructure:"database_name"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`

	//集群配置
	EnableCluster bool   `mapstructure:"enable_cluster"`
	MasterName    string `mapstructure:"master_name"`
	//连接配置
	UseSSL                bool          `mapstructure:"use_ssl"`
	SSLInsecureSkipVerify bool          `mapstructure:"ssl_insecure_skip_verify"`
	PoolSize              int           `mapstructure:"pool_size"`
	MaxActiveConns        int           `mapstructure:"max_active_conns"`
	MinIdleConns          int           `mapstructure:"min_idle_conns"`
	MaxIdleConns          int           `mapstructure:"max_idle_conns"`
	ConnMaxIdleTime       time.Duration `mapstructure:"conn_max_idle_time"`
	ConnMaxLifetime       time.Duration `mapstructure:"conn_max_lifetime"`
	// 超时配置
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
}

func NewRedisOptions() *RedisOptions {
	return &RedisOptions{
		Host:                  "127.0.0.1",
		Port:                  6379,
		Addrs:                 []string{},
		Username:              "",
		Password:              "",
		DB:                    0,
		HealthCheckInterval:   5 * time.Second,
		EnableCluster:         false,
		MasterName:            "",
		UseSSL:                false,
		SSLInsecureSkipVerify: false,
		PoolSize:              10,
		MaxActiveConns:        100,
		MinIdleConns:          5,
		MaxIdleConns:          10,
		ConnMaxIdleTime:       5 * time.Minute,
		ConnMaxLifetime:       1 * time.Hour,
		DialTimeout:           5 * time.Second,
		ReadTimeout:           3 * time.Second,
		WriteTimeout:          3 * time.Second,
		PoolTimeout:           4 * time.Second,
	}
}

func (o *RedisOptions) Validate() []error {
	errs := []error{}
	if len(o.Addrs) == 0 && o.Host == "" && o.Port == 0 {
		errs = append(errs, errors.New("addrs or port is required"))
	}
	return errs
}

func LoadEnvVars(v *viper.Viper, prefix string) (err error) {
	prefix = prefix + "."
	// 敏感信息必须通过环境变量设置
	envMappings := map[string]string{
		// Redis
		prefix + "host":                     "IAM_REDIS_HOST",
		prefix + "port":                     "IAM_REDIS_PORT",
		prefix + "addrs":                    "IAM_REDIS_ADDRS",
		prefix + "username":                 "IAM_REDIS_USERNAME",
		prefix + "password":                 "IAM_REDIS_PASSWORD",
		prefix + "db":                       "IAM_REDIS_DB",
		prefix + "health_check_interval":    "IAM_REDIS_HEALTH_CHECK_INTERVAL",
		prefix + "enable_cluster":           "IAM_REDIS_ENABLE_CLUSTER",
		prefix + "master_name":              "IAM_REDIS_MASTER_NAME",
		prefix + "use_ssl":                  "IAM_REDIS_USE_SSL",
		prefix + "ssl_insecure_skip_verify": "IAM_REDIS_SSL_INSECURE",
		prefix + "pool_size":                "IAM_REDIS_POOL_SIZE",
		prefix + "max_active_conns":         "IAM_REDIS_MAX_ACTIVE_CONNS",
		prefix + "min_idle_conns":           "IAM_REDIS_MIN_IDLE_CONNS",
		prefix + "max_idle_conns":           "IAM_REDIS_MAX_IDLE_CONNS",
		prefix + "conn_max_idle_time":       "IAM_REDIS_CONN_MAX_IDLE_TIME",
		prefix + "conn_max_lifetime":        "IAM_REDIS_CONN_MAX_LIFETIME",
		prefix + "dial_timeout":             "IAM_REDIS_DIAL_TIMEOUT",
		prefix + "read_timeout":             "IAM_REDIS_READ_TIMEOUT",
		prefix + "write_timeout":            "IAM_REDIS_WRITE_TIMEOUT",
		prefix + "pool_timeout":             "IAM_REDIS_POOL_TIMEOUT",
	}

	for key, env := range envMappings {
		if err := v.BindEnv(key, env); err != nil {
			fmt.Printf("%s 未绑定: %v\n", key, err)
		}
	}
	return nil
}

func (o *RedisOptions) SetDefaults(v *viper.Viper, prefix string) {
	prefix = prefix + "."
	opts := NewRedisOptions()

	// Redis默认值
	v.SetDefault(prefix+"host", opts.Host)
	v.SetDefault(prefix+"port", opts.Port)
	v.SetDefault(prefix+"addrs", opts.Addrs)
	v.SetDefault(prefix+"username", opts.Username)
	v.SetDefault(prefix+"password", opts.Password)
	v.SetDefault(prefix+"db", opts.DB)
	v.SetDefault(prefix+"health_check_interval", opts.HealthCheckInterval)
	v.SetDefault(prefix+"enable_cluster", opts.EnableCluster)
	v.SetDefault(prefix+"master_name", opts.MasterName)
	v.SetDefault(prefix+"use_ssl", opts.UseSSL)
	v.SetDefault(prefix+"ssl_insecure_skip_verify", opts.SSLInsecureSkipVerify)
	v.SetDefault(prefix+"pool_size", opts.PoolSize)
	v.SetDefault(prefix+"max_active_conns", opts.MaxActiveConns)
	v.SetDefault(prefix+"min_idle_conns", opts.MinIdleConns)
	v.SetDefault(prefix+"max_idle_conns", opts.MaxIdleConns)
	v.SetDefault(prefix+"conn_max_idle_time", opts.ConnMaxIdleTime)
	v.SetDefault(prefix+"conn_max_lifetime", opts.ConnMaxLifetime)
	v.SetDefault(prefix+"dial_timeout", opts.DialTimeout)
	v.SetDefault(prefix+"read_timeout", opts.ReadTimeout)
	v.SetDefault(prefix+"write_timeout", opts.WriteTimeout)
	v.SetDefault(prefix+"pool_timeout", opts.PoolTimeout)
}
