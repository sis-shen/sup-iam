package rpc

import (
	"context"
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service"
	pbv1 "github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v1"
	"github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v2"
	"strconv"
	"time"
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
