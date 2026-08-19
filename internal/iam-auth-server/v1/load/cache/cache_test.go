package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/dgraph-io/ristretto"
	cachedmodel "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/rpc"
	"github.com/stretchr/testify/require"
)

// mockRpcClient 用于测试的 rpc 客户端，可模拟重载失败
type mockRpcClient struct {
	secretsErr  error
	policiesErr error
}

func (m *mockRpcClient) GetAllSecrets(ctx context.Context) ([]*cachedmodel.CachedSecret, error) {
	return nil, m.secretsErr
}

func (m *mockRpcClient) GetAllPolicies(ctx context.Context) ([][]*cachedmodel.CachedPolicy, error) {
	return nil, m.policiesErr
}

var _ rpc.RpcClient = (*mockRpcClient)(nil)

func newTestCache(cli rpc.RpcClient) *Cache {
	return &Cache{
		cli: cli,
		cacheConfig: &ristretto.Config{
			NumCounters: 100,
			MaxCost:     1000,
			BufferItems: 64,
		},
	}
}

func TestCache_ReloadFiresHandlers(t *testing.T) {
	c := newTestCache(&mockRpcClient{})

	calls := 0
	c.RegisterReloadHandler(func() { calls++ })

	require.NoError(t, c.ReloadSecrets())
	require.Equal(t, 1, calls, "secrets reload 成功后应触发回调")

	require.NoError(t, c.ReloadPolicies())
	require.Equal(t, 2, calls, "policies reload 成功后应触发回调")
}

func TestCache_ReloadFailedSkipsHandlers(t *testing.T) {
	c := newTestCache(&mockRpcClient{secretsErr: errors.New("rpc down")})

	calls := 0
	c.RegisterReloadHandler(func() { calls++ })

	require.Error(t, c.ReloadSecrets())
	require.Equal(t, 0, calls, "reload 失败时本地缓存未更新，不应触发回调")

	require.NoError(t, c.ReloadPolicies())
	require.Equal(t, 1, calls)
}

func TestCache_ReloadHandlerRegisteredAfterNotify(t *testing.T) {
	c := newTestCache(&mockRpcClient{})

	// 先触发一次重载（此时尚无回调）
	require.NoError(t, c.ReloadPolicies())

	// 之后再注册，只影响后续重载
	calls := 0
	c.RegisterReloadHandler(func() { calls++ })

	require.NoError(t, c.ReloadPolicies())
	require.Equal(t, 1, calls)
}
