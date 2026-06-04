package options

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	opts := NewOptions()

	assert.NotNil(t, opts)
	assert.NotNil(t, opts.Pumps)
	assert.Contains(t, opts.Pumps, "mongo")

	mongoOpts := opts.Pumps["mongo"]
	assert.Equal(t, "mongo", mongoOpts.Type)
	assert.Equal(t, time.Second, mongoOpts.Timeout)
	assert.Equal(t, true, mongoOpts.OmitDetailRecoding)
	assert.NotNil(t, mongoOpts.Filters)
	assert.Nil(t, mongoOpts.Filters.Usernames)
	assert.Nil(t, mongoOpts.Filters.SkippedUsernames)

	assert.Equal(t, 10*time.Second, opts.LeaderLeaseDuration)
	assert.Equal(t, time.Second, opts.PurgeInterval)
	assert.Equal(t, "/health", opts.HealthCheckPath)
	assert.Equal(t, "0.0.0.0:7070", opts.HealthCheckAddress)
	assert.Equal(t, true, opts.OmitDetailRecoding)
	assert.NotNil(t, opts.RedisOptions)
	assert.NotNil(t, opts.Log)
}

func TestNewOptions_MongoMeta(t *testing.T) {
	opts := NewOptions()
	mongoMeta := opts.Pumps["mongo"].Meta

	assert.Equal(t, "mongodb://localhost:27017", mongoMeta["url"])
	assert.Equal(t, true, mongoMeta["use_ssl"])
	assert.Equal(t, true, mongoMeta["ssl_skip_verify"])
	assert.Equal(t, true, mongoMeta["ssl_allow_invalid_hosts"])
	assert.Equal(t, "/opt/ssl/certs/ca-bundle.pem", mongoMeta["ssl_ca_file"])
	assert.Equal(t, "/opt/ssl/certs/key.pem", mongoMeta["ssl_kem_key_file"])
	assert.Equal(t, "", mongoMeta["db_type"])
	assert.Equal(t, "analytics", mongoMeta["collection_name"])
	assert.Equal(t, 5*1024*1024, mongoMeta["max_insert_batch_size_bytes"])
	assert.Equal(t, 5*1024*1024, mongoMeta["max_document_size_bytes"])
	assert.Equal(t, 5*1024*1024, mongoMeta["collection_cap_max_size_bytes"])
	assert.Equal(t, true, mongoMeta["collection_cap_enabled"])
}

func TestOptions_Validate(t *testing.T) {
	opts := NewOptions()
	errs := opts.Validate()
	// Should still pass validation even with mongo defaults
	assert.Empty(t, errs)
}

func TestPumpOptions_ZeroFilters(t *testing.T) {
	p := PumpOptions{
		Filters: nil,
	}
	assert.Nil(t, p.Filters)
}

func TestNewOptions_DefaultRedisOptions(t *testing.T) {
	opts := NewOptions()
	assert.Equal(t, "127.0.0.1", opts.RedisOptions.Host)
	assert.Equal(t, 6379, opts.RedisOptions.Port)
}
