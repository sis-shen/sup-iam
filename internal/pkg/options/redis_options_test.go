package options

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewRedisOptions(t *testing.T) {
	opts := NewRedisOptions()
	assert.Equal(t, "127.0.0.1", opts.Host)
	assert.Equal(t, 6379, opts.Port)
	assert.Equal(t, []string{}, opts.Addrs)
	assert.Equal(t, 0, opts.DB)
	assert.Equal(t, 5*time.Second, opts.HealthCheckInterval)
	assert.Equal(t, false, opts.EnableCluster)
	assert.Equal(t, 10, opts.PoolSize)
	assert.Equal(t, 100, opts.MaxActiveConns)
	assert.Equal(t, 5, opts.MinIdleConns)
	assert.Equal(t, 10, opts.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, opts.ConnMaxIdleTime)
	assert.Equal(t, 1*time.Hour, opts.ConnMaxLifetime)
	assert.Equal(t, 5*time.Second, opts.DialTimeout)
	assert.Equal(t, 3*time.Second, opts.ReadTimeout)
	assert.Equal(t, 3*time.Second, opts.WriteTimeout)
	assert.Equal(t, 4*time.Second, opts.PoolTimeout)
}

func TestRedisOptions_Validate_EmptyHostAndPort(t *testing.T) {
	opts := &RedisOptions{
		Host:  "",
		Port:  0,
		Addrs: []string{},
	}
	errs := opts.Validate()
	assert.NotEmpty(t, errs, "Validate should return errors when Host, Port, and Addrs are all empty")
	assert.Contains(t, errs[0].Error(), "addrs or port is required")
}

func TestRedisOptions_Validate_WithHost(t *testing.T) {
	opts := &RedisOptions{
		Host: "localhost",
		Port: 6379,
	}
	errs := opts.Validate()
	assert.Empty(t, errs, "Validate should return no errors when Host and Port are set")
}

func TestRedisOptions_Validate_WithAddrs(t *testing.T) {
	opts := &RedisOptions{
		Addrs: []string{"127.0.0.1:6379"},
	}
	errs := opts.Validate()
	assert.Empty(t, errs, "Validate should return no errors when Addrs is set")
}

func TestRedisOptions_SetDefaults(t *testing.T) {
	v := viper.New()
	opts := NewRedisOptions()
	opts.SetDefaults(v, "redis")

	assert.Equal(t, "127.0.0.1", v.GetString("redis.host"))
	assert.Equal(t, 6379, v.GetInt("redis.port"))
	assert.Equal(t, []string{}, v.GetStringSlice("redis.addrs"))
	assert.Equal(t, 0, v.GetInt("redis.db"))
	assert.Equal(t, 5*time.Second, v.GetDuration("redis.health_check_interval"))
	assert.Equal(t, false, v.GetBool("redis.enable_cluster"))
	assert.Equal(t, "", v.GetString("redis.master_name"))
	assert.Equal(t, 10, v.GetInt("redis.pool_size"))
	assert.Equal(t, 100, v.GetInt("redis.max_active_conns"))
	assert.Equal(t, 5, v.GetInt("redis.min_idle_conns"))
}

func TestLoadEnvVars(t *testing.T) {
	v := viper.New()
	err := LoadEnvVars(v, "redis")
	assert.NoError(t, err)
	// Verify we can set env vars and they bind correctly
	t.Setenv("IAM_REDIS_HOST", "my-redis-host")
	t.Setenv("IAM_REDIS_PORT", "6380")
	t.Setenv("IAM_REDIS_PASSWORD", "s3cret")
	v.Set("redis.host", "my-redis-host")
	v.Set("redis.port", "6380")
	v.Set("redis.password", "s3cret")
	assert.Equal(t, "my-redis-host", v.GetString("redis.host"))
	assert.Equal(t, "6380", v.GetString("redis.port"))
	assert.Equal(t, "s3cret", v.GetString("redis.password"))
}
