package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestNewRedisTokenBlackList(t *testing.T) {
	client := newTestRedisClient(t)
	bl := NewRedisTokenBlackList(client, 10*time.Minute)
	assert.NotNil(t, bl)
	assert.Equal(t, client, bl.client)
	assert.Equal(t, 10*time.Minute, bl.ttl)
}

func TestRedisTokenBlackList_Add_Success(t *testing.T) {
	client := newTestRedisClient(t)
	bl := NewRedisTokenBlackList(client, 10*time.Minute)

	err := bl.Add(context.Background(), "test-token-123")
	assert.NoError(t, err)
}

func TestRedisTokenBlackList_IsBlacklisted_True(t *testing.T) {
	client := newTestRedisClient(t)
	bl := NewRedisTokenBlackList(client, 10*time.Minute)

	err := bl.Add(context.Background(), "token-in-blacklist")
	require.NoError(t, err)

	blacklisted, err := bl.IsBlacklisted(context.Background(), "token-in-blacklist")
	assert.NoError(t, err)
	assert.True(t, blacklisted)
}

func TestRedisTokenBlackList_IsBlacklisted_False(t *testing.T) {
	client := newTestRedisClient(t)
	bl := NewRedisTokenBlackList(client, 10*time.Minute)

	blacklisted, err := bl.IsBlacklisted(context.Background(), "non-existent-token")
	assert.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestRedisTokenBlackList_Add_ClientError(t *testing.T) {
	// 先保存 Addr，再 Close — Close 后 Addr() 会 panic (miniredis 内部 server 被置 nil)
	mr, err := miniredis.Run()
	require.NoError(t, err)
	addr := mr.Addr()
	mr.Close()
	client := redis.NewClient(&redis.Options{Addr: addr})
	bl := NewRedisTokenBlackList(client, 10*time.Minute)

	// 使用已取消的 context，保证立即返回错误，不依赖网络超时
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = bl.Add(ctx, "fail-token")
	assert.Error(t, err)
}

func TestRedisTokenBlackList_IsBlacklisted_ClientError(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	addr := mr.Addr()
	mr.Close()
	client := redis.NewClient(&redis.Options{Addr: addr})
	bl := NewRedisTokenBlackList(client, 10*time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	blacklisted, err := bl.IsBlacklisted(ctx, "fail-token")
	assert.Error(t, err)
	assert.False(t, blacklisted)
}

// TestRedisTokenBlackList_ImplementsInterface ensures compile-time interface check
func TestRedisTokenBlackList_ImplementsInterface(t *testing.T) {
	client := newTestRedisClient(t)
	bl := NewRedisTokenBlackList(client, 5*time.Minute)

	var _ TokenBlackListInterface = bl
}
