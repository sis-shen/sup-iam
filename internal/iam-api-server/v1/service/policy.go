package service

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
)

type PolicyCaseInterface interface {
	GetPolicyList(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Policy], error)
	DeletePolicy(ctx context.Context, id string) error
	GetPolicyByID(ctx context.Context, id string) (*model.Policy, error)
	UpdatePolicy(ctx context.Context, policy *model.Policy) (*model.Policy, error)
	GetPolicyBindingSecretList(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Secret], error)
	CreatePolicy(ctx context.Context, policy *model.Policy) (*model.Policy, error)
}

type PolicyCase struct {
	repo repository.PolicyRepository
}

func NewPolicyCase(repo repository.PolicyRepository) *PolicyCase {
	return &PolicyCase{repo: repo}
}

var _ PolicyCaseInterface = (*PolicyCase)(nil)

func (pc *PolicyCase) GetPolicyList(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Policy], error) {
	return pc.repo.GetListByUserID(ctx, id, query)
}

func (pc *PolicyCase) DeletePolicy(ctx context.Context, id string) error {
	return pc.repo.DeleteByID(ctx, id)
}

func (pc *PolicyCase) GetPolicyByID(ctx context.Context, id string) (*model.Policy, error) {
	return pc.repo.GetByID(ctx, id)
}

func (pc *PolicyCase) UpdatePolicy(ctx context.Context, policy *model.Policy) (*model.Policy, error) {
	return pc.repo.Update(ctx, policy)
}

func (pc *PolicyCase) CreatePolicy(ctx context.Context, policy *model.Policy) (*model.Policy, error) {
	return pc.repo.Create(ctx, policy)
}

func (pc *PolicyCase) GetPolicyBindingSecretList(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Secret], error) {
	return pc.repo.GetSecretListByPolicyID(ctx, id, query)
}
