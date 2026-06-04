package redis

import (
	storage "github.com/sis-shen/sup-iam/internal/iam-pump/store"
	"github.com/sis-shen/sup-iam/internal/pkg/options"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAddrs_WithAddrs(t *testing.T) {
	o := &options.RedisOptions{
		Addrs: []string{"10.0.0.1:6379", "10.0.0.2:6379"},
		Port:  6379, // Explicit Port prevents default address append
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"10.0.0.1:6379", "10.0.0.2:6379"}, addrs)
}

func TestGetAddrs_WithAddrsAndZeroPortAppendsDefault(t *testing.T) {
	o := &options.RedisOptions{
		Addrs: []string{"cluster-1:6379"},
		Port:  0,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"cluster-1:6379", "127.0.0.1:6379"}, addrs,
		"should append default address when Addrs is set but Port is 0")
}

func TestGetAddrs_WithHostAndPort(t *testing.T) {
	o := &options.RedisOptions{
		Host: "my-redis.example.com",
		Port: 6380,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"my-redis.example.com:6380"}, addrs)
}

func TestGetAddrs_Empty(t *testing.T) {
	o := &options.RedisOptions{
		Host:  "",
		Port:  0,
		Addrs: nil,
	}
	addrs := getAddrs(o)
	assert.Empty(t, addrs)
}

func TestGetAddrs_EmptyAddrsWithPort(t *testing.T) {
	o := &options.RedisOptions{
		Host:  "redis-host",
		Port:  6379,
		Addrs: []string{},
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"redis-host:6379"}, addrs)
}

func TestGetAddrs_WithHostOnly(t *testing.T) {
	o := &options.RedisOptions{
		Host:  "localhost",
		Port:  0,
		Addrs: nil,
	}
	addrs := getAddrs(o)
	assert.Empty(t, addrs)
}

func TestDefaultRedisAddress(t *testing.T) {
	assert.Equal(t, "127.0.0.1:6379", defaultRedisAddress)
}

func TestRedisKeyPrefix(t *testing.T) {
	assert.Equal(t, "analytics-", RedisKeyPrefix)
}

func TestRedisClusterStorageManager_GetName(t *testing.T) {
	m := &RedisClusterStorageManager{}
	assert.Equal(t, "redis", m.GetName())
}

func TestRedisClusterStorageManager_SetKeyPrefix(t *testing.T) {
	m := &RedisClusterStorageManager{}
	m.SetKeyPrefix("custom-prefix-")
	assert.Equal(t, "custom-prefix-", m.keyPrefix)
}

func TestRedisClusterStorageManager_ImplementsInterface(t *testing.T) {
	// Compile-time check: RedisClusterStorageManager implements storage.AnalyticStoreInterface
	var _ storage.AnalyticStoreInterface = (*RedisClusterStorageManager)(nil)
	assert.True(t, true)
}
