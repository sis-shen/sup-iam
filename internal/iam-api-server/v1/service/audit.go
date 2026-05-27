package service

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
)

type AuditCaseInterface interface {
	GetAuditBindingList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.BindingAudit], error)
	GetAuditBindingByID(ctx context.Context, id string) (model.BindingAudit, error)
	GetAuditPolicyList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.PolicyAudit], error)
	GetAuditPolicyByID(ctx context.Context, id string) (model.PolicyAudit, error)
}

type AuditCase struct {
	repo repository.AuditRepository
}

func NewAuditCase(repo repository.AuditRepository) *AuditCase {
	return &AuditCase{repo}
}

var _ (AuditCaseInterface) = (*AuditCase)(nil)

func (ac *AuditCase) GetAuditBindingList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.BindingAudit], error) {
	return repository.PageResult[*model.BindingAudit]{}, nil
}

func (ac *AuditCase) GetAuditBindingByID(ctx context.Context, id string) (model.BindingAudit, error) {
	return model.BindingAudit{}, nil
}

func (ac *AuditCase) GetAuditPolicyList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.PolicyAudit], error) {
	return repository.PageResult[*model.PolicyAudit]{}, nil
}

func (ac *AuditCase) GetAuditPolicyByID(ctx context.Context, id string) (model.PolicyAudit, error) {
	return model.PolicyAudit{}, nil
}
