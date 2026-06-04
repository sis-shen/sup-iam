package rpc

import (
	"context"
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service"
	"github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v2"
)

type AuthQueryHandler struct {
	pbv2.UnimplementedAuthQueryServiceServer
	secretCase service.SecretCaseInterface
}

func NewAuthQueryHandler(secretCase service.SecretCaseInterface) *AuthQueryHandler {
	return &AuthQueryHandler{
		secretCase: secretCase,
	}
}

var _ pbv2.AuthQueryServiceServer = (*AuthQueryHandler)(nil)

func (a *AuthQueryHandler) GetAllSecrets(ctx context.Context, req *pbv2.GetAllSecretsRequest) (*pbv2.GetAllSecretsResponse, error) {
	secrets, err := a.secretCase.GetAllSecrets(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*pbv2.Secret, 0, len(secrets))
	for _, secret := range secrets {
		result = append(result, &pbv2.Secret{
			SecretId:  fmt.Sprintf("%d", secret.ID),
			SecretKey: secret.SecretKey,
			AccessKey: secret.AccessKey,
			ExpiresAt: secret.Expires,
		})
	}
	return &pbv2.GetAllSecretsResponse{Secrets: result}, nil
}

func (a *AuthQueryHandler) GetAllPolicies(ctx context.Context, req *pbv2.GetAllPoliciesRequest) (*pbv2.GetAllPoliciesResponse, error) {
	policyMap, err := a.secretCase.GetAllPolicies(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*pbv2.PolicyGroup, 0, len(policyMap))
	for secretID, policyList := range policyMap {
		policies := make([]*pbv2.Policy, 0, len(policyList))
		for _, policy := range policyList {
			policies = append(policies, &pbv2.Policy{
				PolicyId:  fmt.Sprintf("%d", policy.ID),
				SecretId:  secretID,
				Username:  policy.Username,
				PolicyDsl: *policy.PolicyShadow,
			})
		}
		result = append(result, &pbv2.PolicyGroup{PolicyGroup: policies})
	}
	return &pbv2.GetAllPoliciesResponse{PolicyGroups: result}, nil
}
