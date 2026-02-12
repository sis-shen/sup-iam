package repository

import "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"

type AuditRepository interface {
	GetPolicyAuditByID(string) (*model.PolicyAudit, error)
	GetPolicyAuditList(query PageQuery) (PageResult[*model.PolicyAudit], error)
	GetBindingAuditByID(string) (*model.BindingAudit, error)
	GetBindingAuditList(query PageQuery) (PageResult[*model.BindingAudit], error)
}
