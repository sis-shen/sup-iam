package repository

import "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"

type SecretRepository interface {
	Create(*model.Secret) (*model.Secret, error)
	GetByID(string) (*model.Secret, error)
	Update(*model.Secret) (*model.Secret, error)
	DeleteByID(string) error
	GetListByUserID(userID string, query PageQuery) (PageResult[*model.Secret], error)
	GetPolicyListBySecretID(secretID string, query PageQuery) (PageResult[*model.Policy], error)
}
