package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	appmodel "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/rpc/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/keys"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// newTestKeys creates a Keys instance suitable for testing
func newTestKeys() keys.Keys {
	return keys.Keys{}
}

// newTestEnforcer creates a Casbin enforcer with a basic RBAC model for testing
func newTestEnforcer(t *testing.T) *casbin.Enforcer {
	t.Helper()
	m, err := casbinmodel.NewModelFromString(`
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
`)
	require.NoError(t, err)
	e, err := casbin.NewEnforcer(m)
	require.NoError(t, err)
	return e
}

// policyShadow is a helper to create a *string for PolicyShadow
func policyShadow(s string) *string {
	return &s
}

func TestNewAuthCase(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)
	require.NotNil(t, ac)
	require.Equal(t, mockCli, ac.cli)
}

func TestBuildCanonicalString(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)

	result := ac.BuildCanonicalString("AK123", "POST", "/api/v1/resource", "hash123", "1700000000")
	expected := "AK123\nPOST\n/api/v1/resource\nhash123\n1700000000"
	require.Equal(t, expected, result)
}

func TestBuildCanonicalString_EmptyFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)

	result := ac.BuildCanonicalString("", "", "", "", "")
	expected := "\n\n\n\n"
	require.Equal(t, expected, result)
}

func TestVerifySecretKey_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)
	k := newTestKeys()

	// Prepare a secret that won't expire
	secretKey := "test-secret-key-for-signing"
	accessKey := "AK-test-001"
	canonicalString := "AK-test-001\nGET\n/resource\nhash\n1700000000"

	// Sign the payload to get the expected signature
	signature, err := k.SignWithKey(secretKey, canonicalString)
	require.NoError(t, err)

	// Set up mock to return a non-expired secret
	mockCli.EXPECT().
		GetSecretByAK(gomock.Any(), accessKey).
		Return(&appmodel.Secret{
			SecretKey: secretKey,
			AccessKey: accessKey,
			Expires:   time.Now().Add(1 * time.Hour).Unix(),
		}, nil)

	ok, secret, err := ac.VerifySecretKey(context.Background(), accessKey, canonicalString, signature)
	require.NoError(t, err)
	require.True(t, ok)
	require.NotNil(t, secret)
	require.Equal(t, accessKey, secret.AccessKey)
}

func TestVerifySecretKey_InvalidSignature(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)

	accessKey := "AK-test-002"
	canonicalString := "AK-test-002\nGET\n/resource\nhash\n1700000000"

	mockCli.EXPECT().
		GetSecretByAK(gomock.Any(), accessKey).
		Return(&appmodel.Secret{
			SecretKey: "real-secret-key",
			AccessKey: accessKey,
			Expires:   time.Now().Add(1 * time.Hour).Unix(),
		}, nil)

	ok, secret, err := ac.VerifySecretKey(context.Background(), accessKey, canonicalString, "invalid-signature")
	require.NoError(t, err)
	require.False(t, ok)
	require.NotNil(t, secret)
}

func TestVerifySecretKey_Expired(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)

	accessKey := "AK-expired"

	mockCli.EXPECT().
		GetSecretByAK(gomock.Any(), accessKey).
		Return(&appmodel.Secret{
			SecretKey: "key",
			AccessKey: accessKey,
			Expires:   time.Now().Add(-1 * time.Hour).Unix(),
		}, nil)

	ok, secret, err := ac.VerifySecretKey(context.Background(), accessKey, "canonical", "signature")
	require.Error(t, err)
	require.Contains(t, err.Error(), "secret expired")
	require.False(t, ok)
	require.Nil(t, secret)
}

func TestVerifySecretKey_RPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)

	mockCli.EXPECT().
		GetSecretByAK(gomock.Any(), "AK-error").
		Return(nil, errors.New("rpc connection failed"))

	ok, secret, err := ac.VerifySecretKey(context.Background(), "AK-error", "canonical", "signature")
	require.Error(t, err)
	require.Contains(t, err.Error(), "rpc connection failed")
	require.False(t, ok)
	require.Nil(t, secret)
}

func TestAuthorize_MatchedPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	e := newTestEnforcer(t)
	ac := &AuthCase{
		cli:      mockCli,
		keys:     keys.Keys{},
		enforcer: *e,
	}

	policyJSON := `[["alice", "/api/resource", "GET"]]`
	mockCli.EXPECT().
		GetPolicyListBySecretID(gomock.Any(), "1").
		Return([]*appmodel.Policy{
			{ID: 100, PolicyShadow: policyShadow(policyJSON)},
		}, nil)

	ok, matched, err := ac.Authorize(context.Background(), "1", "alice", "/api/resource", "GET")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, []string{"100"}, matched)
}

func TestAuthorize_NoMatchingPolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	e := newTestEnforcer(t)
	ac := &AuthCase{
		cli:      mockCli,
		keys:     keys.Keys{},
		enforcer: *e,
	}

	policyJSON := `[["bob", "/api/other", "POST"]]`
	mockCli.EXPECT().
		GetPolicyListBySecretID(gomock.Any(), "2").
		Return([]*appmodel.Policy{
			{ID: 200, PolicyShadow: policyShadow(policyJSON)},
		}, nil)

	ok, matched, err := ac.Authorize(context.Background(), "2", "alice", "/api/resource", "GET")
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, matched)
}

func TestAuthorize_MultiplePolicies_StopsOnFirstMatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	e := newTestEnforcer(t)
	ac := &AuthCase{
		cli:      mockCli,
		keys:     keys.Keys{},
		enforcer: *e,
	}

	mockCli.EXPECT().
		GetPolicyListBySecretID(gomock.Any(), "3").
		Return([]*appmodel.Policy{
			{ID: 1, PolicyShadow: policyShadow(`[["alice", "/api/a", "GET"]]`)},
			{ID: 2, PolicyShadow: policyShadow(`[["alice", "/api/resource", "GET"]]`)},
		}, nil)

	ok, matched, err := ac.Authorize(context.Background(), "3", "alice", "/api/resource", "GET")
	require.NoError(t, err)
	require.True(t, ok)
	// Returns the matching policy (second one, since first has /api/a not /api/resource)
	require.Equal(t, []string{"2"}, matched)
}

func TestAuthorize_RPCError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	e := newTestEnforcer(t)
	ac := &AuthCase{
		cli:      mockCli,
		keys:     keys.Keys{},
		enforcer: *e,
	}

	mockCli.EXPECT().
		GetPolicyListBySecretID(gomock.Any(), "4").
		Return(nil, errors.New("rpc error"))

	ok, matched, err := ac.Authorize(context.Background(), "4", "alice", "/api/resource", "GET")
	require.Error(t, err)
	require.False(t, ok)
	require.Nil(t, matched)
}

func TestAuthorize_InvalidPolicyJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	e := newTestEnforcer(t)
	ac := &AuthCase{
		cli:      mockCli,
		keys:     keys.Keys{},
		enforcer: *e,
	}

	mockCli.EXPECT().
		GetPolicyListBySecretID(gomock.Any(), "5").
		Return([]*appmodel.Policy{
			{ID: 5, PolicyShadow: policyShadow(`not-valid-json`)},
		}, nil)

	ok, matched, err := ac.Authorize(context.Background(), "5", "alice", "/api/resource", "GET")
	require.Error(t, err)
	require.False(t, ok)
	require.Nil(t, matched)
}

func TestAuthorize_NoPolicies(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	e := newTestEnforcer(t)
	ac := &AuthCase{
		cli:      mockCli,
		keys:     keys.Keys{},
		enforcer: *e,
	}

	mockCli.EXPECT().
		GetPolicyListBySecretID(gomock.Any(), "6").
		Return([]*appmodel.Policy{}, nil)

	ok, matched, err := ac.Authorize(context.Background(), "6", "alice", "/api/resource", "GET")
	require.NoError(t, err)
	require.False(t, ok)
	require.Empty(t, matched)
}

func TestAuthCaseInterfaceImplementation(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCli := mock.NewMockRpcClientInterface(ctrl)
	ac := NewAuthCase(mockCli)

	var _ AuthCaseInterface = ac
}
