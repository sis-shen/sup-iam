package analytics

import (
	"fmt"
	"time"
)

type AnalyticsOptions struct {
	PoolSize              int           `mapstructure:"pool_size"`
	RecordBufferSize      int           `mapstructure:"record_buffer_size"`
	FlushInterval         time.Duration `mapstructure:"flush_interval"`
	StorageExpiration     time.Duration `mapstructure:"storage_expiration"`
	Enable                bool          `mapstructure:"enable"`
	EnableDetailRecording bool          `mapstructure:"enable_detail_recording"`
	AnalyticsKeyName      string        `mapstructure:"analytics_key_name"`
}

func NewAnalyticsOptions() *AnalyticsOptions {
	return &AnalyticsOptions{
		PoolSize:              10,
		RecordBufferSize:      1024,
		FlushInterval:         10 * time.Second,
		StorageExpiration:     24 * time.Hour,
		Enable:                true,
		EnableDetailRecording: true,
		AnalyticsKeyName:      "iam-system-analytics",
	}
}

// Validate is used to parse and validate the parameters entered by the user at
// the command line when the program starts.
func (o *AnalyticsOptions) Validate() []error {
	if o == nil {
		return nil
	}
	errors := []error{}

	if o.Enable && (o.FlushInterval < time.Second || o.FlushInterval > 1000*time.Second) {
		errors = append(errors, fmt.Errorf("--analytics.flush-interval %v must be between 1 and 1000", o.FlushInterval))
	}

	return errors
}
