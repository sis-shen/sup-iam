package repository

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
)

type AuditRepository interface {
	GetPolicyAuditByID(ctx context.Context, id string) (*model.PolicyAudit, error)
	GetPolicyAuditList(ctx context.Context, query PageQuery) (PageResult[*model.PolicyAudit], error)
	GetBindingAuditByID(ctx context.Context, id string) (*model.BindingAudit, error)
	GetBindingAuditList(ctx context.Context, query PageQuery) (PageResult[*model.BindingAudit], error)
}
