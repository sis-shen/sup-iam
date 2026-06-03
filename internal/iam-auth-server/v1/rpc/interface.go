package rpc

//go:generate mockgen -destination=./mock/rpc_client_mock.go -package=mock . RpcClientInterface

import (
	"context"
	cachedmodel "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/model"
	pbv2 "github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v2"
	"time"
)

type RpcClient interface {
	GetAllSecrets(ctx context.Context) ([]*cachedmodel.CachedSecret, error)
	GetAllPolicies(ctx context.Context) ([][]*cachedmodel.CachedPolicy, error)
}

type GRpcClient struct {
	cli pbv2.AuthQueryServiceClient
}

func NewGRpcClient(cli pbv2.AuthQueryServiceClient) *GRpcClient {
	return &GRpcClient{
		cli: cli,
	}
}

var _ RpcClient = (*GRpcClient)(nil)

func (c *GRpcClient) GetAllSecrets(ctx context.Context) ([]*cachedmodel.CachedSecret, error) {
	resp, err := c.cli.GetAllSecrets(ctx, &pbv2.GetAllSecretsRequest{})
	if err != nil {
		return nil, err
	}
	result := make([]*cachedmodel.CachedSecret, 0, len(resp.Secrets))
	for _, secret := range resp.Secrets {
		cached := &cachedmodel.CachedSecret{
			AccessKey: secret.AccessKey,
			SecretKey: secret.SecretKey,
			ExpiredAt: time.Unix(secret.ExpiresAt, 0),
			ID:        secret.SecretId,
		}
		result = append(result, cached)
	}
	return result, nil
}

func (c *GRpcClient) GetAllPolicies(ctx context.Context) ([][]*cachedmodel.CachedPolicy, error) {
	resp, err := c.cli.GetAllPolicies(ctx, &pbv2.GetAllPoliciesRequest{})
	if err != nil {
		return nil, err
	}
	result := make([][]*cachedmodel.CachedPolicy, 0, len(resp.PolicyGroups))
	for _, group := range resp.PolicyGroups {
		policyGroup := group.PolicyGroup
		if len(policyGroup) > 0 {
			policies := make([]*cachedmodel.CachedPolicy, 0, len(policyGroup))
			for _, policy := range policyGroup {
				policies = append(policies, &cachedmodel.CachedPolicy{
					SecretID: policy.SecretId,
					Username: policy.Username,
					ID:       policy.PolicyId,
					DSL:      policy.PolicyDsl,
				})
			}
			result = append(result, policies)
		}
	}

	return result, nil
}
