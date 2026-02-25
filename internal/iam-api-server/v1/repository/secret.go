package repository

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
)

type SecretRepository interface {
	Create(ctx context.Context, secret *model.Secret) (*model.Secret, error)
	GetByID(ctx context.Context, id string) (*model.Secret, error)
	Update(ctx context.Context, secret *model.Secret) (*model.Secret, error)
	DeleteByID(ctx context.Context, id string) error
	GetListByUserID(ctx context.Context, id string, query PageQuery) (PageResult[*model.Secret], error)
	GetPolicyListBySecretID(ctx context.Context, id string, query PageQuery) (PageResult[*model.Policy], error)
}
