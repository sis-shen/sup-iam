package rpc

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service"
	"github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v1"
	"strconv"
	"time"
)

type AuthQueryHandler struct {
	pbv1.UnimplementedAuthQueryServiceServer
	secretCase service.SecretCaseInterface
}

func NewAuthQueryHandler(secretCase service.SecretCaseInterface) *AuthQueryHandler {
	return &AuthQueryHandler{
		secretCase: secretCase,
	}
}

var _ pbv1.AuthQueryServiceServer = (*AuthQueryHandler)(nil)

func (a *AuthQueryHandler) GetSecretByAK(ctx context.Context, in *pbv1.GetSecretByAKRequest) (*pbv1.GetSecretByAKResponse, error) {
	secret, err := a.secretCase.GetSecretByAK(ctx, in.GetAccessKey())
	if err != nil {
		return nil, err
	}

	expiresAt := secret.UpdatedAt.Add(time.Duration(secret.Expires) * time.Second)
	return &pbv1.GetSecretByAKResponse{
		Secret: &pbv1.Secret{
			SecretId:  strconv.FormatUint(secret.ID, 10),
			SecretKey: secret.SecretKey,
			AccessKey: secret.AccessKey,
			ExpiresAt: expiresAt.Unix(),
		},
	}, nil
}

func (a *AuthQueryHandler) GetPolicyBySecretID(ctx context.Context, in *pbv1.GetPolicyBySecretIDRequest) (*pbv1.GetPolicyBySecretIDResponse, error) {
	query := repository.PageQuery{
		Limit:   100,
		Cursor:  "",
		OrderBy: "",
		Order:   "",
	}

	id := in.SecretId

	res, err := a.secretCase.GetSecretBindingPolicy(ctx, id, query)
	if err != nil {
		return nil, err
	}

	policyList := make([]*pbv1.Policy, 0)

	for _, policy := range res.Items {
		if policy.PolicyShadow == nil {
			continue
		}
		p := &pbv1.Policy{
			PolicyId:  strconv.FormatUint(policy.ID, 10),
			Username:  policy.Username,
			PolicyDsl: *policy.PolicyShadow,
		}
		policyList = append(policyList, p)
	}

	return &pbv1.GetPolicyBySecretIDResponse{
		PolicyList: policyList,
	}, nil
}
