package analytics

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewAnalyticsOptions(t *testing.T) {
	opts := NewAnalyticsOptions()
	assert.Equal(t, 10, opts.PoolSize)
	assert.Equal(t, 1024, opts.RecordBufferSize)
	assert.Equal(t, 10*time.Second, opts.FlushInterval)
	assert.Equal(t, 24*time.Hour, opts.StorageExpiration)
	assert.Equal(t, true, opts.Enable)
	assert.Equal(t, true, opts.EnableDetailRecording)
	assert.Equal(t, "iam-system-analytics", opts.AnalyticsKeyName)
}

func TestAnalyticsOptions_Validate_NilReceiver(t *testing.T) {
	var opts *AnalyticsOptions
	errs := opts.Validate()
	assert.Nil(t, errs, "Validate on nil receiver should return nil")
}

func TestAnalyticsOptions_Validate_Disabled(t *testing.T) {
	opts := &AnalyticsOptions{
		Enable: false,
	}
	errs := opts.Validate()
	assert.Empty(t, errs, "Validate on disabled analytics should return no errors")
}

func TestAnalyticsOptions_Validate_EnabledWithValidInterval(t *testing.T) {
	opts := &AnalyticsOptions{
		Enable:        true,
		FlushInterval: 10 * time.Second,
	}
	errs := opts.Validate()
	assert.Empty(t, errs, "Validate with valid flush interval (10s) should return no errors")
}

func TestAnalyticsOptions_Validate_EnabledWithOneSecondInterval(t *testing.T) {
	opts := &AnalyticsOptions{
		Enable:        true,
		FlushInterval: 1 * time.Second,
	}
	errs := opts.Validate()
	assert.Empty(t, errs, "Validate with 1s flush interval should return no errors")
}

func TestAnalyticsOptions_Validate_EnabledWithSmallInterval(t *testing.T) {
	opts := &AnalyticsOptions{
		Enable:        true,
		FlushInterval: 500 * time.Millisecond,
	}
	errs := opts.Validate()
	assert.NotEmpty(t, errs, "Validate with flush interval < 1s should return error")
	assert.Contains(t, errs[0].Error(), "flush-interval")
}

func TestAnalyticsOptions_Validate_EnabledWithLargeInterval(t *testing.T) {
	opts := &AnalyticsOptions{
		Enable:        true,
		FlushInterval: 2000 * time.Second,
	}
	errs := opts.Validate()
	assert.NotEmpty(t, errs, "Validate with flush interval > 1000s should return error")
	assert.Contains(t, errs[0].Error(), "flush-interval")
}

func TestAnalyticsOptions_Validate_EnabledWithZeroInterval(t *testing.T) {
	opts := &AnalyticsOptions{
		Enable:        true,
		FlushInterval: 0,
	}
	errs := opts.Validate()
	assert.NotEmpty(t, errs, "Validate with zero flush interval should return error")
}
