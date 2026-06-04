package options

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"github.com/spf13/viper"
)

type PumpOptions struct {
	Type               string                     `mapstructure:"type"`
	Filters            *analytics.AnalyticFilters `mapstructure:"filters"`
	Timeout            time.Duration              `mapstructure:"timeout"`
	OmitDetailRecoding bool                       `mapstructure:"omit_detail_recoding"`
	Meta               map[string]interface{}     `mapstructure:"meta"`
}

type Options struct {
	Pumps               map[string]PumpOptions       `mapstructure:"pumps"`
	LeaderLeaseDuration time.Duration                `mapstructure:"leader_lease_duration"`
	PurgeInterval       time.Duration                `mapstructure:"purge_interval"`
	HealthCheckPath     string                       `mapstructure:"health_check_path"`
	HealthCheckAddress  string                       `mapstructure:"health_check_address"`
	OmitDetailRecoding  bool                         `mapstructure:"omit_detail_recoding"`
	RedisOptions        *genericoptions.RedisOptions `mapstructure:"redis_options"`
	Log                 *log.Options                 `mapstructure:"log"`
}

func NewOptions() *Options {
	s := &Options{
		Pumps: map[string]PumpOptions{
			"mongo": {
				Type: "mongo",
				Filters: &analytics.AnalyticFilters{
					Usernames:        nil,
					SkippedUsernames: nil,
				},
				Timeout:            time.Second,
				OmitDetailRecoding: true,
				Meta: map[string]interface{}{
					"url":                           "mongodb://localhost:27017",
					"use_ssl":                       true,
					"ssl_skip_verify":               true,
					"ssl_allow_invalid_hosts":       true,
					"ssl_ca_file":                   "/opt/ssl/certs/ca-bundle.pem",
					"ssl_kem_key_file":              "/opt/ssl/certs/key.pem",
					"db_type":                       "",
					"collection_name":               "analytics",
					"max_insert_batch_size_bytes":   5 * 1024 * 1024,
					"max_document_size_bytes":       5 * 1024 * 1024,
					"collection_cap_max_size_bytes": 5 * 1024 * 1024,
					"collection_cap_enabled":        true,
				},
			},
		},
		LeaderLeaseDuration: 10 * time.Second,
		PurgeInterval:       time.Second,
		HealthCheckPath:     "/health",
		HealthCheckAddress:  "0.0.0.0:7070",
		OmitDetailRecoding:  true,
		RedisOptions:        genericoptions.NewRedisOptions(),
		Log:                 log.NewOptions(),
	}
	return s
}

func (o *Options) Validate() []error {
	errs := []error{}
	errs = append(errs, o.RedisOptions.Validate()...)
	errs = append(errs, o.Log.Validate()...)
	return errs
}

func Load(configPath string) *Options {
	//不支持环境变量
	return LoadConfigFile(configPath)
}

func LoadConfigFile(path string) *Options {
	if path == "" {
		path = autoConfigFilePath()
	}
	if path == "" {
		log.Fatalf("未找到配置文件，使用默认值，当前程序不支持环境变量")
	}
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("读取配置文件失败")
	}

	cfg := NewOptions()
	if err := viper.Unmarshal(cfg); err != nil {
		log.Errorf("解析配置文件失败:%w", err)
		return nil
	}

	return cfg

}

func autoConfigFilePath() string {

	// 1. 从环境变量获取
	if envFile := os.Getenv("IAM_PUMP_CONFIG_FILE"); envFile != "" {
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
