package options

import (
	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"time"
)

type PumpOptions struct {
	Type               string                    `mapstructure:"type"`
	Filters            analytics.AnalyticFilters `mapstructure:"filters"`
	Timeout            time.Duration             `mapstructure:"timeout"`
	OmitDetailRecoding bool                      `mapstructure:"omit_detail_recoding"`
	Meta               map[string]interface{}    `mapstructure:"meta"`
}

type Options struct {
	Pumps               map[string]PumpOptions       `mapstructure:"pumps"`
	LeaderLeaseDuration time.Duration                `mapstructure:"leader_lease_duration"`
	PurgeInterval       time.Duration                `mapstructure:"purge_interval"`
	HealthCheckPath     string                       `mapstructure:"health_check_path"`
	HealthCheckAddress  string                       `mapstructure:"health_check_interval"`
	OmitDetailRecoding  bool                         `mapstructure:"omit_detail_recoding"`
	RedisOptions        *genericoptions.RedisOptions `mapstructure:"redis_options"`
	Log                 *log.Options                 `mapstructure:"log"`
}

func NewOptions() *Options {
	s := &Options{
		Pumps: map[string]PumpOptions{
			"mongo": {
				Type: "mongo",
				Filters: analytics.AnalyticFilters{
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
