package rpc

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v1"
	"strconv"
)

type RpcClientInterface interface {
	GetSecretByAK(ctx context.Context, accessKey string) (*model.Secret, error)
	GetPolicyListBySecretID(ctx context.Context, secretID string) ([]*model.Policy, error)
}

type GRpcClient struct {
	cli pbv1.AuthQueryServiceClient
}

func NewGRpcClient(cli pbv1.AuthQueryServiceClient) *GRpcClient {
	return &GRpcClient{
		cli: cli,
	}
}

var _ RpcClientInterface = (*GRpcClient)(nil)

func (c *GRpcClient) GetSecretByAK(ctx context.Context, accessKey string) (*model.Secret, error) {
	req := pbv1.GetSecretByAKRequest{
		AccessKey: accessKey,
	}
	resp, err := c.cli.GetSecretByAK(ctx, &req)
	if err != nil {
		return nil, err
	}
	secret := resp.GetSecret()
	return &model.Secret{
		SecretKey: secret.GetSecretKey(),
		AccessKey: secret.GetAccessKey(),
		Expires:   secret.GetExpiresAt(),
	}, nil
}

func (c *GRpcClient) GetPolicyListBySecretID(ctx context.Context, secretID string) ([]*model.Policy, error) {
	req := pbv1.GetPolicyBySecretIDRequest{
		SecretId: secretID,
	}
	resp, err := c.cli.GetPolicyBySecretID(ctx, &req)
	if err != nil {
		return nil, err
	}
	lst := resp.GetPolicyList()
	res := make([]*model.Policy, len(lst))
	for i, policy := range lst {
		id, err := strconv.ParseUint(policy.GetPolicyId(), 10, 64)
		if err != nil {
			return nil, err
		}

		dsl := policy.GetPolicyDsl()
		res[i] = &model.Policy{
			ID:           id,
			Username:     policy.GetUsername(),
			PolicyShadow: &dsl,
		}
	}
	return res, nil
}
