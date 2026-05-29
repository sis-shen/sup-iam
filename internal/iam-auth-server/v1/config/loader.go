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
		"server.host":          "IAM_SERVER_HOST",
		"server.port":          "IAM_SERVER_PORT",
		"server.mode":          "IAM_SERVER_MODE",
		"server.read_timeout":  "IAM_SERVER_READ_TIMEOUT",
		"server.write_timeout": "IAM_SERVER_WRITE_TIMEOUT",
		"server.idle_timeout":  "IAM_SERVER_IDLE_TIMEOUT",
		"server.grace_timeout": "IAM_SERVER_GRACE_TIMEOUT",

		// gRPC
		"grpc.host": "IAM_GRPC_HOST",
		"grpc.port": "IAM_GRPC_PORT",

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

	// gRPC默认值
	v.SetDefault("grpc.host", "127.0.0.1")
	v.SetDefault("grpc.port", 9090)

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
