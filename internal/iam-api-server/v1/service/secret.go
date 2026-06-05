package service

//go:generate mockgen -destination=./mock/secret_case_mock.go -package=mock . SecretCaseInterface

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/sis-shen/sup-iam/internal/pkg/keys"
	"time"
)

type SecretCaseInterface interface {
	GetSecretList(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Secret], error)
	DeleteSecret(ctx context.Context, id string) error
	GetSecretByID(ctx context.Context, id string) (*model.Secret, error)
	GetSecretByAK(ctx context.Context, ak string) (*model.Secret, error)
	GetSecretBindingPolicy(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Policy], error)
	UpdateSecret(ctx context.Context, model *model.Secret) (*model.Secret, error)
	CreateSecret(ctx context.Context, secret *model.Secret) (*model.Secret, error)
	RotateSecret(ctx context.Context, id string) (*model.Secret, error)
	VerifySecret(secret string, payload string, hashedStr string) (bool, error)
	GenerateSecretKey() string
	GenerateAccessKey() string
	GetAllSecrets(ctx context.Context) ([]*model.Secret, error)
	GetAllPolicies(ctx context.Context) (map[string][]*model.Policy, error)
}

type SecretCase struct {
	repo repository.SecretRepository
	keys keys.KeysInterface
}

var _ SecretCaseInterface = (*SecretCase)(nil)

func NewSecretCase(repo repository.SecretRepository) *SecretCase {
	return &SecretCase{repo: repo, keys: keys.NewKeys(0, 0)}
}

func (sc *SecretCase) GetSecretList(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Secret], error) {
	return sc.repo.GetListByUserID(ctx, id, query)
}

func (sc *SecretCase) DeleteSecret(ctx context.Context, id string) error {
	return sc.repo.DeleteByID(ctx, id)
}

func (sc *SecretCase) GetSecretByID(ctx context.Context, id string) (*model.Secret, error) {
	return sc.repo.GetByID(ctx, id)
}

func (sc *SecretCase) GetSecretByAK(ctx context.Context, ak string) (*model.Secret, error) {
	return sc.repo.GetByAK(ctx, ak)
}

func (sc *SecretCase) CreateSecret(ctx context.Context, secret *model.Secret) (*model.Secret, error) {
	return sc.repo.Create(ctx, secret)
}

func (sc *SecretCase) UpdateSecret(ctx context.Context, model *model.Secret) (*model.Secret, error) {
	return sc.repo.Update(ctx, model)
}

func (sc *SecretCase) RotateSecret(ctx context.Context, id string) (*model.Secret, error) {
	secret, err := sc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	secret.SecretKey = sc.keys.GenerateSecretKey()
	nowTime := time.Now()
	secret.UpdatedAt = nowTime
	secret, err = sc.repo.Update(ctx, secret)
	return secret, err
}

func (sc *SecretCase) VerifySecret(secret string, payload string, hashedStr string) (bool, error) {
	return sc.keys.VerifySecretKey(secret, payload, hashedStr)
}

func (sc *SecretCase) GenerateSecretKey() string {
	return sc.keys.GenerateSecretKey()
}

func (sc *SecretCase) GenerateAccessKey() string {
	return sc.keys.GenerateAccessKey()
}

func (sc *SecretCase) GetSecretBindingPolicy(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Policy], error) {
	return sc.repo.GetPolicyListBySecretID(ctx, id, query)
}

func (sc *SecretCase) GetAllSecrets(ctx context.Context) ([]*model.Secret, error) {
	return sc.repo.GetAllSecrets(ctx)
}
func (sc *SecretCase) GetAllPolicies(ctx context.Context) (map[string][]*model.Policy, error) {
	return sc.repo.GetAllPoliciesMap(ctx)
}
