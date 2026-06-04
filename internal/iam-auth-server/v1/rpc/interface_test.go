package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/model"
	pbv2 "github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

// mockAuthQueryServiceClient implements pbv2.AuthQueryServiceClient for testing
type mockAuthQueryServiceClient struct {
	getAllSecretsFunc  func(ctx context.Context, in *pbv2.GetAllSecretsRequest, opts ...grpc.CallOption) (*pbv2.GetAllSecretsResponse, error)
	getAllPoliciesFunc func(ctx context.Context, in *pbv2.GetAllPoliciesRequest, opts ...grpc.CallOption) (*pbv2.GetAllPoliciesResponse, error)
}

func (m *mockAuthQueryServiceClient) GetAllSecrets(ctx context.Context, in *pbv2.GetAllSecretsRequest, opts ...grpc.CallOption) (*pbv2.GetAllSecretsResponse, error) {
	return m.getAllSecretsFunc(ctx, in, opts...)
}

func (m *mockAuthQueryServiceClient) GetAllPolicies(ctx context.Context, in *pbv2.GetAllPoliciesRequest, opts ...grpc.CallOption) (*pbv2.GetAllPoliciesResponse, error) {
	return m.getAllPoliciesFunc(ctx, in, opts...)
}

func TestNewGRpcClient(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{}
	client := NewGRpcClient(mockCli)
	assert.NotNil(t, client)
	assert.Equal(t, mockCli, client.cli)
}

func TestGRpcClient_ImplementsInterface(t *testing.T) {
	var _ RpcClient = (*GRpcClient)(nil)
}

func TestGRpcClient_GetAllSecrets_Success(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllSecretsFunc: func(ctx context.Context, in *pbv2.GetAllSecretsRequest, opts ...grpc.CallOption) (*pbv2.GetAllSecretsResponse, error) {
			return &pbv2.GetAllSecretsResponse{
				Secrets: []*pbv2.Secret{
					{
						SecretId:  "secret-1",
						SecretKey: "sk-abc123",
						AccessKey: "ak-xyz789",
						ExpiresAt: 0, // never expires
					},
					{
						SecretId:  "secret-2",
						SecretKey: "sk-def456",
						AccessKey: "ak-uvw012",
						ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
					},
				},
			}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	secrets, err := client.GetAllSecrets(context.Background())
	assert.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.Equal(t, "secret-1", secrets[0].ID)
	assert.Equal(t, "sk-abc123", secrets[0].SecretKey)
	assert.Equal(t, "ak-xyz789", secrets[0].AccessKey)
	// ExpiresAt=0 is epoch time (1970-01-01), not a zero time.Time
	assert.Equal(t, time.Unix(0, 0), secrets[0].ExpiredAt)

	assert.Equal(t, "secret-2", secrets[1].ID)
	assert.Equal(t, "sk-def456", secrets[1].SecretKey)
	assert.Equal(t, "ak-uvw012", secrets[1].AccessKey)
	assert.False(t, secrets[1].ExpiredAt.IsZero())
}

func TestGRpcClient_GetAllSecrets_Empty(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllSecretsFunc: func(ctx context.Context, in *pbv2.GetAllSecretsRequest, opts ...grpc.CallOption) (*pbv2.GetAllSecretsResponse, error) {
			return &pbv2.GetAllSecretsResponse{
				Secrets: []*pbv2.Secret{},
			}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	secrets, err := client.GetAllSecrets(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, secrets)
}

func TestGRpcClient_GetAllSecrets_Error(t *testing.T) {
	expectedErr := errors.New("rpc error: connection refused")
	mockCli := &mockAuthQueryServiceClient{
		getAllSecretsFunc: func(ctx context.Context, in *pbv2.GetAllSecretsRequest, opts ...grpc.CallOption) (*pbv2.GetAllSecretsResponse, error) {
			return nil, expectedErr
		},
	}

	client := NewGRpcClient(mockCli)
	secrets, err := client.GetAllSecrets(context.Background())
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, secrets)
}

func TestGRpcClient_GetAllSecrets_NilSecrets(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllSecretsFunc: func(ctx context.Context, in *pbv2.GetAllSecretsRequest, opts ...grpc.CallOption) (*pbv2.GetAllSecretsResponse, error) {
			return &pbv2.GetAllSecretsResponse{}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	secrets, err := client.GetAllSecrets(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, secrets)
}

func TestGRpcClient_GetAllPolicies_Success(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllPoliciesFunc: func(ctx context.Context, in *pbv2.GetAllPoliciesRequest, opts ...grpc.CallOption) (*pbv2.GetAllPoliciesResponse, error) {
			return &pbv2.GetAllPoliciesResponse{
				PolicyGroups: []*pbv2.PolicyGroup{
					{
						PolicyGroup: []*pbv2.Policy{
							{
								PolicyId:  "policy-1",
								SecretId:  "secret-1",
								Username:  "alice",
								PolicyDsl: `{"effect":"allow"}`,
							},
							{
								PolicyId:  "policy-2",
								SecretId:  "secret-1",
								Username:  "alice",
								PolicyDsl: `{"effect":"deny"}`,
							},
						},
					},
					{
						PolicyGroup: []*pbv2.Policy{
							{
								PolicyId:  "policy-3",
								SecretId:  "secret-2",
								Username:  "bob",
								PolicyDsl: `{"effect":"allow"}`,
							},
						},
					},
				},
			}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	policies, err := client.GetAllPolicies(context.Background())
	assert.NoError(t, err)
	assert.Len(t, policies, 2)
	assert.Len(t, policies[0], 2)
	assert.Len(t, policies[1], 1)

	assert.Equal(t, "policy-1", policies[0][0].ID)
	assert.Equal(t, "secret-1", policies[0][0].SecretID)
	assert.Equal(t, "alice", policies[0][0].Username)
	assert.Equal(t, `{"effect":"allow"}`, policies[0][0].DSL)
}

func TestGRpcClient_GetAllPolicies_Empty(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllPoliciesFunc: func(ctx context.Context, in *pbv2.GetAllPoliciesRequest, opts ...grpc.CallOption) (*pbv2.GetAllPoliciesResponse, error) {
			return &pbv2.GetAllPoliciesResponse{
				PolicyGroups: []*pbv2.PolicyGroup{},
			}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	policies, err := client.GetAllPolicies(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, policies)
}

func TestGRpcClient_GetAllPolicies_Error(t *testing.T) {
	expectedErr := errors.New("rpc error: timeout")
	mockCli := &mockAuthQueryServiceClient{
		getAllPoliciesFunc: func(ctx context.Context, in *pbv2.GetAllPoliciesRequest, opts ...grpc.CallOption) (*pbv2.GetAllPoliciesResponse, error) {
			return nil, expectedErr
		},
	}

	client := NewGRpcClient(mockCli)
	policies, err := client.GetAllPolicies(context.Background())
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, policies)
}

func TestGRpcClient_GetAllPolicies_EmptyPolicyGroup(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllPoliciesFunc: func(ctx context.Context, in *pbv2.GetAllPoliciesRequest, opts ...grpc.CallOption) (*pbv2.GetAllPoliciesResponse, error) {
			return &pbv2.GetAllPoliciesResponse{
				PolicyGroups: []*pbv2.PolicyGroup{
					{}, // empty group should be skipped
					{
						PolicyGroup: []*pbv2.Policy{
							{
								PolicyId:  "policy-1",
								SecretId:  "secret-1",
								Username:  "alice",
								PolicyDsl: `{"effect":"allow"}`,
							},
						},
					},
				},
			}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	policies, err := client.GetAllPolicies(context.Background())
	assert.NoError(t, err)
	assert.Len(t, policies, 1, "empty policy group should be filtered out")
	assert.Len(t, policies[0], 1)
}

func TestGRpcClient_GetAllPolicies_NilResponse(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllPoliciesFunc: func(ctx context.Context, in *pbv2.GetAllPoliciesRequest, opts ...grpc.CallOption) (*pbv2.GetAllPoliciesResponse, error) {
			return &pbv2.GetAllPoliciesResponse{}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	policies, err := client.GetAllPolicies(context.Background())
	assert.NoError(t, err)
	assert.Empty(t, policies)
}

func TestCachedModel_Mapping(t *testing.T) {
	mockCli := &mockAuthQueryServiceClient{
		getAllSecretsFunc: func(ctx context.Context, in *pbv2.GetAllSecretsRequest, opts ...grpc.CallOption) (*pbv2.GetAllSecretsResponse, error) {
			return &pbv2.GetAllSecretsResponse{
				Secrets: []*pbv2.Secret{
					{
						SecretId:  "s-1",
						SecretKey: "key",
						AccessKey: "ak",
						ExpiresAt: 0,
					},
				},
			}, nil
		},
	}

	client := NewGRpcClient(mockCli)
	secrets, err := client.GetAllSecrets(context.Background())
	assert.NoError(t, err)

	// Verify correct type mapping
	assert.IsType(t, &model.CachedSecret{}, secrets[0])
	assert.Equal(t, "s-1", secrets[0].ID)
	assert.Equal(t, "key", secrets[0].SecretKey)
	assert.Equal(t, "ak", secrets[0].AccessKey)
}
