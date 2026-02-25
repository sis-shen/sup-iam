package repository

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
)

type PolicyRepository interface {
	Create(ctx context.Context, policy *model.Policy) (*model.Policy, error)
	GetByID(ctx context.Context, id string) (*model.Policy, error)
	Update(ctx context.Context, policy *model.Policy) (*model.Policy, error)
	DeleteByID(ctx context.Context, id string) error
	GetListByUserID(ctx context.Context, id string, query PageQuery) (PageResult[*model.Policy], error)
	GetSecretListByPolicyID(ctx context.Context, id string, query PageQuery) (PageResult[*model.Secret], error)
}
