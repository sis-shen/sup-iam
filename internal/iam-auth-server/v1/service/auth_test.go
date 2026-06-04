package service

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/analytics"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/load/cache"
	cachedmodel "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/storage"
	"github.com/sis-shen/sup-iam/internal/pkg/keys"
	"github.com/stretchr/testify/require"
)

// mockRpcClient implements rpc.RpcClient for testing
type mockRpcClient struct {
	mu       sync.Mutex
	secrets  []*cachedmodel.CachedSecret
	policies [][]*cachedmodel.CachedPolicy
}

func (m *mockRpcClient) GetAllSecrets(ctx context.Context) ([]*cachedmodel.CachedSecret, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.secrets, nil
}

func (m *mockRpcClient) GetAllPolicies(ctx context.Context) ([][]*cachedmodel.CachedPolicy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.policies, nil
}

// mockAnalyticsStore implements storage.AnalyticsStore for testing
type mockAnalyticsStore struct{}

func (m *mockAnalyticsStore) Connect() error                                               { return nil }
func (m *mockAnalyticsStore) AppendToSetPipelined(string, [][]byte) error                  { return nil }
func (m *mockAnalyticsStore) SetExpire(string, time.Duration) error                        { return nil }
func (m *mockAnalyticsStore) GetExpire(string) (time.Duration, error)                      { return 0, nil }
func (m *mockAnalyticsStore) SetKeyPrefix(string)                                          {}
func (m *mockAnalyticsStore) WithStopChan(stopChan <-chan struct{}) storage.AnalyticsStore { return m }
func (m *mockAnalyticsStore) WithExpireTime(d time.Duration) storage.AnalyticsStore        { return m }

// test helpers
var (
	testMockRPC   *mockRpcClient
	testCache     *cache.Cache
	testAnalytics *analytics.Analytics
	testKeys      keys.KeysInterface
)

func TestMain(m *testing.M) {
	// 初始化测试依赖（仅执行一次）
	testMockRPC = &mockRpcClient{}
	testKeys = keys.NewKeys(32, 128)

	var err error
	testCache, err = cache.InitSingleton(testMockRPC, &cache.Options{
		NumCounters: 100,
		MaxCost:     1000,
		BufferItems: 64,
	})
	if err != nil {
		os.Exit(1)
	}

	testAnalytics = analytics.NewAnalytics(
		analytics.NewAnalyticsOptions(),
		&mockAnalyticsStore{},
	)

	os.Exit(m.Run())
}

// newTestAuthCase creates a fresh AuthCase for testing
func newTestAuthCase() *AuthCase {
	return NewAuthCase(testCache, testKeys, testAnalytics)
}

func TestNewAuthCase(t *testing.T) {
	ac := newTestAuthCase()
	require.NotNil(t, ac)
}

func TestBuildCanonicalString(t *testing.T) {
	ac := newTestAuthCase()

	result := ac.BuildCanonicalString("AK123", "POST", "/api/v1/resource", "hash123", "1700000000")
	expected := "AK123\nPOST\n/api/v1/resource\nhash123\n1700000000"
	require.Equal(t, expected, result)
}

func TestBuildCanonicalString_EmptyFields(t *testing.T) {
	ac := newTestAuthCase()

	result := ac.BuildCanonicalString("", "", "", "", "")
	expected := "\n\n\n\n"
	require.Equal(t, expected, result)
}

func setTestSecrets(secrets []*cachedmodel.CachedSecret) {
	testMockRPC.mu.Lock()
	testMockRPC.secrets = secrets
	testMockRPC.mu.Unlock()
	_ = testCache.ReloadSecrets()
}

func setTestPolicies(policies [][]*cachedmodel.CachedPolicy) {
	testMockRPC.mu.Lock()
	testMockRPC.policies = policies
	testMockRPC.mu.Unlock()
	_ = testCache.ReloadPolicies()
}

func TestVerifySecretKey_Success(t *testing.T) {
	secretKey := "test-secret-key-for-signing"
	accessKey := "AK-test-001"
	canonicalString := "AK-test-001\nGET\n/resource\nhash\n1700000000"

	signature, err := testKeys.SignWithKey(secretKey, canonicalString)
	require.NoError(t, err)

	setTestSecrets([]*cachedmodel.CachedSecret{
		{
			AccessKey: accessKey,
			SecretKey: secretKey,
			ID:        "1",
			ExpiredAt: time.Now().Add(1 * time.Hour),
		},
	})

	ac := newTestAuthCase()
	ok, secret, err := ac.VerifySecretKey(accessKey, canonicalString, signature)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, secret)
	require.Equal(t, accessKey, secret.AccessKey)
}

func TestVerifySecretKey_InvalidSignature(t *testing.T) {
	accessKey := "AK-test-002"

	setTestSecrets([]*cachedmodel.CachedSecret{
		{
			AccessKey: accessKey,
			SecretKey: "real-secret-key",
			ID:        "2",
			ExpiredAt: time.Now().Add(1 * time.Hour),
		},
	})

	ac := newTestAuthCase()
	ok, secret, err := ac.VerifySecretKey(accessKey, "canonical-string", "invalid-signature")
	require.NoError(t, err)
	require.False(t, ok)
	require.NotNil(t, secret)
}

func TestVerifySecretKey_Expired(t *testing.T) {
	accessKey := "AK-expired"

	setTestSecrets([]*cachedmodel.CachedSecret{
		{
			AccessKey: accessKey,
			SecretKey: "key",
			ID:        "3",
			ExpiredAt: time.Now().Add(-1 * time.Hour),
		},
	})

	ac := newTestAuthCase()
	ok, secret, err := ac.VerifySecretKey(accessKey, "canonical", "signature")
	require.Error(t, err)
	require.Contains(t, err.Error(), "secret expired")
	require.False(t, ok)
	require.Nil(t, secret)
}

func TestVerifySecretKey_CacheMiss(t *testing.T) {
	// 清空缓存
	setTestSecrets(nil)

	ac := newTestAuthCase()
	ok, secret, err := ac.VerifySecretKey("AK-not-found", "canonical", "signature")
	require.Error(t, err)
	require.False(t, ok)
	require.Nil(t, secret)
}

func TestAuthorize_MatchedPolicy(t *testing.T) {
	setTestPolicies([][]*cachedmodel.CachedPolicy{
		{
			{
				ID:       "100",
				SecretID: "1",
				Username: "alice",
				DSL:      `[["alice", "/api/resource", "GET"]]`,
			},
		},
	})

	ac := newTestAuthCase()
	ok, matched, err := ac.Authorize("1", "alice", "/api/resource", "GET")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []string{"100"}, matched)
}

func TestAuthorize_NoMatchingPolicy(t *testing.T) {
	setTestPolicies([][]*cachedmodel.CachedPolicy{
		{
			{
				ID:       "200",
				SecretID: "2",
				Username: "bob",
				DSL:      `[["bob", "/api/other", "POST"]]`,
			},
		},
	})

	ac := newTestAuthCase()
	ok, matched, err := ac.Authorize("2", "alice", "/api/resource", "GET")
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, matched)
}

func TestAuthorize_FirstMatch(t *testing.T) {
	setTestPolicies([][]*cachedmodel.CachedPolicy{
		{
			{
				ID:       "1",
				SecretID: "3",
				Username: "alice",
				DSL:      `[["alice", "/api/a", "GET"]]`,
			},
			{
				ID:       "2",
				SecretID: "3",
				Username: "alice",
				DSL:      `[["alice", "/api/resource", "GET"]]`,
			},
		},
	})

	ac := newTestAuthCase()
	ok, matched, err := ac.Authorize("3", "alice", "/api/resource", "GET")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []string{"2"}, matched)
}

func TestAuthorize_CacheMiss(t *testing.T) {
	setTestPolicies(nil)

	ac := newTestAuthCase()
	ok, matched, err := ac.Authorize("not-found", "alice", "/api/resource", "GET")
	require.Error(t, err)
	require.False(t, ok)
	require.Nil(t, matched)
}

func TestAuthorize_InvalidPolicyJSON(t *testing.T) {
	setTestPolicies([][]*cachedmodel.CachedPolicy{
		{
			{
				ID:       "5",
				SecretID: "5",
				Username: "alice",
				DSL:      `not-valid-json`,
			},
		},
	})

	ac := newTestAuthCase()
	ok, matched, err := ac.Authorize("5", "alice", "/api/resource", "GET")
	require.Error(t, err)
	require.False(t, ok)
	require.Nil(t, matched)
}

func TestAuthorize_NoPolicies(t *testing.T) {
	// 空策略组会被 ReloadPolicies 跳过，缓存中无对应 key
	setTestPolicies(nil)

	ac := newTestAuthCase()
	ok, matched, err := ac.Authorize("6", "alice", "/api/resource", "GET")
	require.Error(t, err)
	require.False(t, ok)
	require.Nil(t, matched)
}
