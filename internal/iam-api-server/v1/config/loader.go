package config

import (
	"fmt"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func Load(configPath string) (*AppConfig, error) {
	v := viper.New()

	// 1. 设置默认值
	cfg := NewConfig()

	// 2.导入配置文件
	if configPath == "" {
		configPath = autoConfigFilePath()
	}
	if err := LoadConfigFile(v, configPath); err != nil {
		fmt.Println("读取配置文件失败，跳过配置文件")
	}

	// 2. 绑定环境变量
	v.SetEnvPrefix("IAM")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 显式绑定重要环境变量
	LoadEnvVars(v)
	_ = genericoptions.LoadEnvVars(v, "redis")

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 5. 验证配置
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
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
		fmt.Printf("成功加载加载配置文件: %s\n", v.ConfigFileUsed())
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
		"server.host": "IAM_SERVER_HOST",
		"server.port": "IAM_SERVER_PORT",
		"server.mode": "IAM_SERVER_MODE",

		// JWT
		"jwt.secret_key":  "IAM_JWT_SECRET_KEY",
		"jwt.user_id_key": "IAM_JWT_USER_ID_KEY",

		// MySQL
		"mysql.host":     "IAM_MYSQL_HOST",
		"mysql.port":     "IAM_MYSQL_PORT",
		"mysql.username": "IAM_MYSQL_USERNAME",
		"mysql.password": "IAM_MYSQL_PASSWORD",

		// gRPC
		"grpc.host": "IAM_GRPC_HOST",
		"grpc.port": "IAM_GRPC_PORT",

		// 日志
		"log.level": "IAM_LOG_LEVEL",
		"log.name":  "IAM_LOG_NAME",
	}

	for key, env := range envMappings {
		if err := v.BindEnv(key, env); err != nil {
			fmt.Printf("%s 未绑定: %v\n", key, err)
		}
	}
}

func NewConfig() *AppConfig {
	return &AppConfig{
		Server: &ServerConfig{
			Host:              "0.0.0.0",
			Port:              9000,
			Mode:              "debug",
			ReadTimeout:       time.Second * 15,
			WriteTimeout:      time.Second * 15,
			IdleTimeout:       time.Second * 60,
			BlackListTTL:      time.Second,
			GraceTimeout:      time.Second * 60,
			EnableRedisSink:   true,
			RedisLogKeyPrefix: "iam-log",
			SinkLevel:         "info",
		},
		JWT: &JWTConfig{
			SecretKey:              "yourSecretKey",
			AccessTokenExpireTime:  time.Hour,
			RefreshTokenExpireTime: 10 * time.Hour,
			UserIDKey:              "user_id",
			TokenLookup:            "header:Authorization",
			Issuer:                 "https://iam.supdriver.ibm.fake",
			SkipPaths:              nil,
		},
		MySQL: &MySQLConfig{
			Host:            "127.0.0.1",
			Port:            3306,
			Username:        "root",
			Password:        "root",
			DatabaseName:    "iam",
			MaxIdleConns:    10,
			MaxOpenConns:    20,
			ConnMaxLifetime: time.Hour,
			MaxRetries:      10,
		},
		GrpcConfig: &GrpcConfig{
			Host:                "0.0.0.0",
			Port:                9000,
			EtcdServerDiscovery: false,
			EtcdHost:            "127.0.0.1",
			EtcdPort:            9000,
			ServiceName:         "iam",
			LeaseTTL:            time.Hour,
			ServiceAddress:      "127.0.0.1:9000",
		},
		Redis: genericoptions.NewRedisOptions(),
		Log:   log.NewOptions(),
	}
}

func autoConfigFilePath() string {

	// 1. 从环境变量获取
	if envFile := os.Getenv("IAM_CONFIG_FILE"); envFile != "" {
		fmt.Printf("从环境变量IAM_CONFIG_FILE获取到配置路径: %s\n", envFile)
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
	fmt.Printf("选取配置文件名为: %s\n", configName)
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

	fmt.Printf("未找到配置文件")
	return ""
}

func validateConfig(cfg *AppConfig) error {
	// 检查必填字段
	if cfg.JWT.SecretKey == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}

	if cfg.MySQL.Password == "" {
		return fmt.Errorf("MySQL密码不能为空")
	}

	// 验证端口范围
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("服务器端口无效: %d", cfg.Server.Port)
	}

	if cfg.MySQL.Port <= 0 || cfg.MySQL.Port > 65535 {
		return fmt.Errorf("MySQL端口无效: %d", cfg.MySQL.Port)
	}

	if cfg.Redis.Port <= 0 || cfg.Redis.Port > 65535 {
		return fmt.Errorf("Redis端口无效: %d", cfg.Redis.Port)
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
