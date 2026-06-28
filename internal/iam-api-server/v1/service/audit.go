package service

//go:generate mockgen -destination=./mock/audit_case_mock.go -package=mock . AuditCaseInterface

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
	return ac.repo.GetBindingAuditList(ctx, query)
}

func (ac *AuditCase) GetAuditBindingByID(ctx context.Context, id string) (model.BindingAudit, error) {
	result, err := ac.repo.GetBindingAuditByID(ctx, id)
	if err != nil {
		return model.BindingAudit{}, err
	}
	return *result, nil
}

func (ac *AuditCase) GetAuditPolicyList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.PolicyAudit], error) {
	return ac.repo.GetPolicyAuditList(ctx, query)
}

func (ac *AuditCase) GetAuditPolicyByID(ctx context.Context, id string) (model.PolicyAudit, error) {
	result, err := ac.repo.GetPolicyAuditByID(ctx, id)
	if err != nil {
		return model.PolicyAudit{}, err
	}
	return *result, nil
}
