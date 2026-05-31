package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 1. 设置默认值
	setDefaults(v)

	// 2. 导入配置文件
	if configPath == "" {
		configPath = autoConfigFilePath()
	}
	if err := LoadConfigFile(v, configPath); err != nil {
		fmt.Println("读取配置文件失败，跳过配置文件")
	}

	// 3. 绑定环境变量
	v.SetEnvPrefix("IAM")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 显式绑定重要环境变量
	LoadEnvVars(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 4. 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

func LoadConfigFile(v *viper.Viper, configPath string) (err error) {
	if configPath == "" {
		configPath = autoConfigFilePath()
	}
	configFile := configPath

	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
		fmt.Printf("成功加载配置文件: %s\n", v.ConfigFileUsed())
	} else {
		fmt.Println("未找到配置文件，使用默认值和环境变量")
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	return nil
}

func LoadEnvVars(v *viper.Viper) {
	// 敏感信息必须通过环境变量设置
	envMappings := map[string]string{
		// 服务器
		"server.host":              "IAM_SERVER_HOST",
		"server.port":              "IAM_SERVER_PORT",
		"server.mode":              "IAM_SERVER_MODE",
		"server.read_timeout":      "IAM_SERVER_READ_TIMEOUT",
		"server.write_timeout":     "IAM_SERVER_WRITE_TIMEOUT",
		"server.idle_timeout":      "IAM_SERVER_IDLE_TIMEOUT",
		"server.grace_timeout":     "IAM_SERVER_GRACE_TIMEOUT",
		"server.enable_redis_sink": "IAM_SERVER_ENABLE_REDIS_SINK",
		"server.redis_key_prefix":  "IAM_SERVER_REDIS_KEY_PREFIX",
		"server.sink_level":        "IAM_SERVER_SINK_LEVEL",

		// gRPC
		"grpc.host":                  "IAM_GRPC_HOST",
		"grpc.port":                  "IAM_GRPC_PORT",
		"grpc.etcd_server_discovery": "IAM_GRPC_ETCD_SERVER_DISCOVERY",
		"grpc.service_name":          "IAM_GRPC_SERVICE_NAME",

		// Redis
		"redis.host":                  "IAM_REDIS_HOST",
		"redis.port":                  "IAM_REDIS_PORT",
		"redis.password":              "IAM_REDIS_PASSWORD",
		"redis.database_name":         "IAM_REDIS_DATABASE_NAME",
		"redis.health_check_interval": "IAM_REDIS_HEALTH_CHECK_INTERVAL",
		"redis.pool_size":             "IAM_REDIS_POOL_SIZE",
		"redis.min_idle_conns":        "IAM_REDIS_MIN_IDLE_CONNS",
		"redis.max_idle_conns":        "IAM_REDIS_MAX_IDLE_CONNS",
		"redis.conn_max_idle_time":    "IAM_REDIS_CONN_MAX_IDLE_TIME",
		"redis.conn_max_lifetime":     "IAM_REDIS_CONN_MAX_LIFETIME",
		"redis.dial_timeout":          "IAM_REDIS_DIAL_TIMEOUT",
		"redis.read_timeout":          "IAM_REDIS_READ_TIMEOUT",
		"redis.write_timeout":         "IAM_REDIS_WRITE_TIMEOUT",
		"redis.pool_timeout":          "IAM_REDIS_POOL_TIMEOUT",

		// 日志
		"log.level":              "IAM_LOG_LEVEL",
		"log.format":             "IAM_LOG_FORMAT",
		"log.output-paths":       "IAM_LOG_OUTPUT_PATHS",
		"log.error-output-paths": "IAM_LOG_ERROR_OUTPUT_PATHS",
		"log.disable-caller":     "IAM_LOG_DISABLE_CALLER",
		"log.disable-stacktrace": "IAM_LOG_DISABLE_STACKTRACE",
		"log.enable-color":       "IAM_LOG_ENABLE_COLOR",
		"log.development":        "IAM_LOG_DEVELOPMENT",
		"log.name":               "IAM_LOG_NAME",
	}

	for key, env := range envMappings {
		if err := v.BindEnv(key, env); err != nil {
			fmt.Printf("%s 未绑定: %v\n", key, err)
		}
	}
}

func setDefaults(v *viper.Viper) {
	// 服务器默认值
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8889)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "120s")
	v.SetDefault("server.grace_timeout", "10s")
	v.SetDefault("server.enable_redis_sink", false)
	v.SetDefault("server.redis_key_prefix", "iam:log:")
	v.SetDefault("server.sink_level", "")

	// gRPC默认值
	v.SetDefault("grpc.host", "127.0.0.1")
	v.SetDefault("grpc.port", 9090)
	v.SetDefault("grpc.etcd_server_discovery", false)
	v.SetDefault("grpc.service_name", "")

	// Redis默认值
	v.SetDefault("redis.host", "127.0.0.1")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.database_name", 0)
	v.SetDefault("redis.health_check_interval", "10s")
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.max_idle_conns", 10)
	v.SetDefault("redis.conn_max_idle_time", "5m")
	v.SetDefault("redis.conn_max_lifetime", "1h")
	v.SetDefault("redis.dial_timeout", "5s")
	v.SetDefault("redis.read_timeout", "3s")
	v.SetDefault("redis.write_timeout", "3s")
	v.SetDefault("redis.pool_timeout", "4s")

	// 日志默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.output-paths", []string{"stdout"})
	v.SetDefault("log.error-output-paths", []string{"stderr"})
	v.SetDefault("log.disable-caller", false)
	v.SetDefault("log.disable-stacktrace", false)
	v.SetDefault("log.enable-color", false)
	v.SetDefault("log.development", false)
	v.SetDefault("log.name", "")
}

func autoConfigFilePath() string {

	// 1. 从环境变量获取
	if envFile := os.Getenv("IAM_CONFIG_FILE"); envFile != "" {
		return envFile
	}

	// 2. 根据运行环境自动选择
	env := os.Getenv("IAM_ENV")
	if env == "" {
		env = os.Getenv("GO_ENV")
	}

	var configName string
	switch env {
	case "production", "prod":
		configName = "config.prod.yaml"
	case "development", "dev":
		configName = "config.dev.yaml"
	case "test":
		configName = "config.test.yaml"
	default:
		configName = "config.yaml"
	}

	// 3. 搜索配置文件
	searchPaths := []string{
		".",                           // 当前目录
		"config",                      // config目录
		filepath.Join("..", "config"), // 上级的config目录
		"/etc/iam/",                   // 系统配置目录
	}

	for _, path := range searchPaths {
		fullPath := filepath.Join(path, configName)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath
		}
	}

	return ""
}

func validateConfig(cfg *Config) error {
	// 验证端口范围
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("服务器端口无效: %d", cfg.Server.Port)
	}

	if cfg.Grpc.Port <= 0 || cfg.Grpc.Port > 65535 {
		return fmt.Errorf("gRPC端口无效: %d", cfg.Grpc.Port)
	}

	// 验证运行模式
	validModes := map[string]bool{
		"debug":   true,
		"release": true,
		"test":    true,
	}
	if !validModes[cfg.Server.Mode] {
		return fmt.Errorf("无效的运行模式: %s", cfg.Server.Mode)
	}

	return nil
}
