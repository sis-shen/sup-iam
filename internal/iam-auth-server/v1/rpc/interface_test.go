package rpc

import (
	"context"
	"errors"
	"testing"

	pbv1 "github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// simpleMockClient implements pbv1.AuthQueryServiceClient with configurable results
type simpleMockClient struct {
	getSecretByAKFn       func(ctx context.Context, in *pbv1.GetSecretByAKRequest, opts ...grpc.CallOption) (*pbv1.GetSecretByAKResponse, error)
	getPolicyBySecretIDFn func(ctx context.Context, in *pbv1.GetPolicyBySecretIDRequest, opts ...grpc.CallOption) (*pbv1.GetPolicyBySecretIDResponse, error)
}

func (m *simpleMockClient) GetSecretByAK(ctx context.Context, in *pbv1.GetSecretByAKRequest, opts ...grpc.CallOption) (*pbv1.GetSecretByAKResponse, error) {
	return m.getSecretByAKFn(ctx, in, opts...)
}

func (m *simpleMockClient) GetPolicyBySecretID(ctx context.Context, in *pbv1.GetPolicyBySecretIDRequest, opts ...grpc.CallOption) (*pbv1.GetPolicyBySecretIDResponse, error) {
	return m.getPolicyBySecretIDFn(ctx, in, opts...)
}

func TestNewGRpcClient(t *testing.T) {
	client := NewGRpcClient(&simpleMockClient{})
	require.NotNil(t, client)
}

func TestGetSecretByAK_Success(t *testing.T) {
	mockCli := &simpleMockClient{
		getSecretByAKFn: func(ctx context.Context, in *pbv1.GetSecretByAKRequest, opts ...grpc.CallOption) (*pbv1.GetSecretByAKResponse, error) {
			require.Equal(t, "ak-001", in.GetAccessKey())
			return &pbv1.GetSecretByAKResponse{
				Secret: &pbv1.Secret{
					SecretKey: "sk-001",
					AccessKey: "ak-001",
					ExpiresAt: 1700000000,
				},
			}, nil
		},
	}
	client := NewGRpcClient(mockCli)

	secret, err := client.GetSecretByAK(context.Background(), "ak-001")
	require.NoError(t, err)
	require.NotNil(t, secret)
	require.Equal(t, "sk-001", secret.SecretKey)
	require.Equal(t, "ak-001", secret.AccessKey)
	require.Equal(t, int64(1700000000), secret.Expires)
}

func TestGetSecretByAK_EmptyResponse(t *testing.T) {
	mockCli := &simpleMockClient{
		getSecretByAKFn: func(ctx context.Context, in *pbv1.GetSecretByAKRequest, opts ...grpc.CallOption) (*pbv1.GetSecretByAKResponse, error) {
			return &pbv1.GetSecretByAKResponse{}, nil
		},
	}
	client := NewGRpcClient(mockCli)

	secret, err := client.GetSecretByAK(context.Background(), "ak-empty")
	require.NoError(t, err)
	require.NotNil(t, secret)
	require.Empty(t, secret.SecretKey)
	require.Empty(t, secret.AccessKey)
	require.Equal(t, int64(0), secret.Expires)
}

func TestGetSecretByAK_GRPCError(t *testing.T) {
	mockCli := &simpleMockClient{
		getSecretByAKFn: func(ctx context.Context, in *pbv1.GetSecretByAKRequest, opts ...grpc.CallOption) (*pbv1.GetSecretByAKResponse, error) {
			return nil, errors.New("gRPC connection error")
		},
	}
	client := NewGRpcClient(mockCli)

	secret, err := client.GetSecretByAK(context.Background(), "ak-error")
	require.Error(t, err)
	require.Contains(t, err.Error(), "gRPC connection error")
	require.Nil(t, secret)
}

func TestGetPolicyListBySecretID_Success(t *testing.T) {
	policyDSL := `[["alice", "/api/resource", "GET"]]`
	mockCli := &simpleMockClient{
		getPolicyBySecretIDFn: func(ctx context.Context, in *pbv1.GetPolicyBySecretIDRequest, opts ...grpc.CallOption) (*pbv1.GetPolicyBySecretIDResponse, error) {
			require.Equal(t, "secret-1", in.GetSecretId())
			return &pbv1.GetPolicyBySecretIDResponse{
				PolicyList: []*pbv1.Policy{
					{PolicyId: "100", Username: "alice", PolicyDsl: policyDSL},
					{PolicyId: "101", Username: "bob", PolicyDsl: `[["bob", "/api/other", "POST"]]`},
				},
			}, nil
		},
	}
	client := NewGRpcClient(mockCli)

	policies, err := client.GetPolicyListBySecretID(context.Background(), "secret-1")
	require.NoError(t, err)
	require.Len(t, policies, 2)

	require.Equal(t, uint64(100), policies[0].ID)
	require.Equal(t, "alice", policies[0].Username)
	require.Equal(t, policyDSL, *policies[0].PolicyShadow)

	require.Equal(t, uint64(101), policies[1].ID)
	require.Equal(t, "bob", policies[1].Username)
}

func TestGetPolicyListBySecretID_EmptyList(t *testing.T) {
	mockCli := &simpleMockClient{
		getPolicyBySecretIDFn: func(ctx context.Context, in *pbv1.GetPolicyBySecretIDRequest, opts ...grpc.CallOption) (*pbv1.GetPolicyBySecretIDResponse, error) {
			return &pbv1.GetPolicyBySecretIDResponse{}, nil
		},
	}
	client := NewGRpcClient(mockCli)

	policies, err := client.GetPolicyListBySecretID(context.Background(), "secret-empty")
	require.NoError(t, err)
	require.Empty(t, policies)
}

func TestGetPolicyListBySecretID_GRPCError(t *testing.T) {
	mockCli := &simpleMockClient{
		getPolicyBySecretIDFn: func(ctx context.Context, in *pbv1.GetPolicyBySecretIDRequest, opts ...grpc.CallOption) (*pbv1.GetPolicyBySecretIDResponse, error) {
			return nil, errors.New("gRPC error")
		},
	}
	client := NewGRpcClient(mockCli)

	policies, err := client.GetPolicyListBySecretID(context.Background(), "secret-error")
	require.Error(t, err)
	require.Contains(t, err.Error(), "gRPC error")
	require.Nil(t, policies)
}

func TestGetPolicyListBySecretID_InvalidPolicyID(t *testing.T) {
	mockCli := &simpleMockClient{
		getPolicyBySecretIDFn: func(ctx context.Context, in *pbv1.GetPolicyBySecretIDRequest, opts ...grpc.CallOption) (*pbv1.GetPolicyBySecretIDResponse, error) {
			return &pbv1.GetPolicyBySecretIDResponse{
				PolicyList: []*pbv1.Policy{
					{PolicyId: "not-a-number", Username: "alice", PolicyDsl: "[]"},
				},
			}, nil
		},
	}
	client := NewGRpcClient(mockCli)

	policies, err := client.GetPolicyListBySecretID(context.Background(), "secret-bad-id")
	require.Error(t, err)
	require.Nil(t, policies)
}

func TestRpcClientInterfaceImplementation(t *testing.T) {
	client := NewGRpcClient(&simpleMockClient{})
	var _ RpcClientInterface = client
}
