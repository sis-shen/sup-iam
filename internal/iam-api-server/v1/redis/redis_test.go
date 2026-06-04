package redis

import (
	"github.com/sis-shen/sup-iam/internal/pkg/options"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetAddrs_WithAddrs(t *testing.T) {
	o := &options.RedisOptions{
		Addrs: []string{"10.0.0.1:6379", "10.0.0.2:6379"},
		Host:  "ignored-host",
		Port:  9999,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"10.0.0.1:6379", "10.0.0.2:6379"}, addrs)
}

func TestGetAddrs_WithHostAndPort(t *testing.T) {
	o := &options.RedisOptions{
		Host: "my-redis.example.com",
		Port: 6380,
	}
	addrs := getAddrs(o)
	assert.Equal(t, []string{"my-redis.example.com:6380"}, addrs)
}

func TestGetAddrs_WithEmptyAddrsAndZeroPort(t *testing.T) {
	o := &options.RedisOptions{
		Host: "",
		Port: 0,
	}
	addrs := getAddrs(o)
	assert.Empty(t, addrs, "should return empty slice when no addresses configured")
}

func TestGetAddrs_WithAddrsAndZeroPortAppendsDefault(t *testing.T) {
	o := &options.RedisOptions{
		Host:  "ignored",
		Addrs: []string{"cluster-1:6379"},
		Port:  0,
	}
	addrs := getAddrs(o)
	// BUG(?): When Addrs is set but Port is 0, the code appends defaultRedisAddress
	assert.Equal(t, []string{"cluster-1:6379", "127.0.0.1:6379"}, addrs,
		"defaultRedisAddress is appended as fallback when Port is 0")
}

func TestGetAddrs_WithOnlyHostNoPortNoAddrs(t *testing.T) {
	o := &options.RedisOptions{
		Host:  "localhost",
		Port:  0,
		Addrs: []string{},
	}
	addrs := getAddrs(o)
	assert.Empty(t, addrs, "should return empty when only Host is set without Port")
}

func TestGetAddrs_AddrsNonEmptyAndPortNonZero(t *testing.T) {
	o := &options.RedisOptions{
		Host:  "host-a",
		Addrs: []string{"addr-1:6379"},
		Port:  1234,
	}
	addrs := getAddrs(o)
	// When Addrs is non-empty, it takes precedence regardless of Port
	assert.Equal(t, []string{"addr-1:6379"}, addrs)
}

func TestGetRedisClusterSingleton_NilBeforeInit(t *testing.T) {
	// Save old singleton and restore after test
	oldSingleton := redisClusterSingleton
	defer func() {
		redisClusterSingleton = oldSingleton
	}()

	redisClusterSingleton = nil
	singleton := GetRedisClusterSingleton()
	assert.Nil(t, singleton)
}
