package repository

import "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"

type PolicyRepository interface {
	Create(*model.Policy) (*model.Policy, error)
	GetByID(string) (*model.Policy, error)
	Update(*model.Policy) (*model.Policy, error)
	DeleteByID(string) error
	GetListByUserID(userID string, query PageQuery) (PageResult[*model.Policy], error)
	GetSecretListByPolicyID(policyID string, query PageQuery) (PageResult[*model.Secret], error)
}
