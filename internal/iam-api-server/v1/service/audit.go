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
