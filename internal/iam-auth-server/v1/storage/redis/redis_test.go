package redis

import (
	"testing"
	"time"

	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/storage"
	genericoptions "github.com/sis-shen/sup-iam/internal/pkg/options"
	"github.com/stretchr/testify/assert"
)

func TestRedisKeyPrefixConstant(t *testing.T) {
	assert.Equal(t, "analytics-", RedisKeyPrefix)
}

func TestDefaultRedisAddressConstant(t *testing.T) {
	assert.Equal(t, "127.0.0.1:6379", defaultRedisAddress)
}

func TestNewRedisClusterStorage(t *testing.T) {
	r := NewRedisClusterStorage()
	assert.NotNil(t, r)
	assert.Nil(t, r.db)
	assert.Empty(t, r.keyPrefix)
}

func TestRedisClusterStorage_ImplementsInterface(t *testing.T) {
	var _ storage.AnalyticsStore = (*RedisClusterStorage)(nil)
}

func TestGetAddrs_WithAddrs(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Addrs: []string{"10.0.0.1:6379", "10.0.0.2:6379"},
		Host:  "ignored-host",
		Port:  9999,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"10.0.0.1:6379", "10.0.0.2:6379"}, addrs)
}

func TestGetAddrs_WithHostAndPort(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host: "my-redis.example.com",
		Port: 6380,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"my-redis.example.com:6380"}, addrs)
}

func TestGetAddrs_WithEmptyAddrsAndZeroPort(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host: "",
		Port: 0,
	}
	addrs := getAddrs(o)
	assert.Empty(t, addrs, "should return empty slice when no addresses configured")
}

func TestGetAddrs_WithAddrsAndZeroPortAppendsDefault(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host:  "ignored",
		Addrs: []string{"cluster-1:6379"},
		Port:  0,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"cluster-1:6379", "127.0.0.1:6379"}, addrs,
		"defaultRedisAddress is appended as fallback when Port is 0")
}

func TestGetAddrs_WithOnlyHostNoPortNoAddrs(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host:  "localhost",
		Port:  0,
		Addrs: []string{},
	}
	addrs := getAddrs(o)
	assert.Empty(t, addrs, "should return empty when only Host is set without Port")
}

func TestGetAddrs_AddrsNonEmptyAndPortNonZero(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host:  "host-a",
		Addrs: []string{"addr-1:6379"},
		Port:  1234,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"addr-1:6379"}, addrs)
}

func TestGetAddrs_OnlyHostWithPort(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host:  "redis-host",
		Port:  6379,
		Addrs: []string{},
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"redis-host:6379"}, addrs)
}

func TestGetAddrs_NilAddrs(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host:  "",
		Port:  0,
		Addrs: nil,
	}
	addrs := getAddrs(o)
	assert.Empty(t, addrs)
}

func TestGetAddrs_NilAddrsWithHostAndPort(t *testing.T) {
	o := &genericoptions.RedisOptions{
		Host:  "cluster.local",
		Port:  7000,
		Addrs: nil,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"cluster.local:7000"}, addrs)
}

func TestRedisClusterStorage_SetKeyPrefix(t *testing.T) {
	r := NewRedisClusterStorage()
	r.SetKeyPrefix("custom-prefix-")
	assert.Equal(t, "custom-prefix-", r.keyPrefix)
}

func TestRedisClusterStorage_fixKey(t *testing.T) {
	r := NewRedisClusterStorage()
	r.SetKeyPrefix("analytics-")
	result := r.fixKey("my-key")
	assert.Equal(t, "analytics-my-key", result)
}

func TestRedisClusterStorage_fixKey_EmptyPrefix(t *testing.T) {
	r := NewRedisClusterStorage()
	result := r.fixKey("my-key")
	assert.Equal(t, "my-key", result, "empty prefix should not modify key")
}

func TestRedisClusterStorage_fixKey_EmptyKey(t *testing.T) {
	r := NewRedisClusterStorage()
	r.SetKeyPrefix("prefix-")
	result := r.fixKey("")
	assert.Equal(t, "prefix-", result)
}

func TestRedisClusterStorage_WithStopChan(t *testing.T) {
	r := NewRedisClusterStorage()
	stopChan := make(chan struct{})
	result := r.WithStopChan(stopChan)
	assert.Same(t, r, result, "WithStopChan should return the same instance")
	// stopChan is <-chan struct{} in the struct (receive-only), so verify by sending
	assert.NotPanics(t, func() {
		select {
		case stopChan <- struct{}{}:
		default:
		}
	})
}

func TestRedisClusterStorage_WithExpireTime(t *testing.T) {
	r := NewRedisClusterStorage()
	expire := 24 * time.Hour
	result := r.WithExpireTime(expire)
	assert.Same(t, r, result, "WithExpireTime should return the same instance")
	assert.Equal(t, expire, r.expireTime)
}

func TestRedisClusterStorage_WithExpireTime_Zero(t *testing.T) {
	r := NewRedisClusterStorage()
	result := r.WithExpireTime(0)
	assert.Same(t, r, result)
	assert.Equal(t, time.Duration(0), r.expireTime)
}

func TestGetRedisClusterSingleton_NilBeforeInit(t *testing.T) {
	oldSingleton := redisClusterSingleton
	redisClusterSingleton = nil
	defer func() {
		redisClusterSingleton = oldSingleton
	}()

	singleton := GetRedisClusterSingleton()
	assert.Nil(t, singleton)
}
