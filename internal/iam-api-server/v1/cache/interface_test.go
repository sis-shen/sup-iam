package cache

import (
	"context"
	"errors"
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
	// 使用已关闭的客户端模拟 Redis 错误
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	bl := NewRedisTokenBlackList(client, 10*time.Minute)

	// 快速超时避免测试太久
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := bl.Add(ctx, "fail-token")
	assert.Error(t, err)
}

func TestRedisTokenBlackList_IsBlacklisted_ClientError(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	bl := NewRedisTokenBlackList(client, 10*time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	blacklisted, err := bl.IsBlacklisted(ctx, "fail-token")
	assert.Error(t, err)
	assert.False(t, blacklisted)
}

// TestRedisTokenBlackList_ImplementsInterface 编译时接口断言
func TestRedisTokenBlackList_ImplementsInterface(t *testing.T) {
	client := newTestRedisClient(t)
	bl := NewRedisTokenBlackList(client, 5*time.Minute)

	var _ TokenBlackListInterface = bl

	// 验证 Add 和 IsBlacklisted 的签名（编译时检查）
	_ = errors.New // 确保 errors 包被使用（用于类型一致性）
}
