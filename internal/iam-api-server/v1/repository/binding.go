package repository

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
)

type BindingRepository interface {
	Create(ctx context.Context, binding *model.Binding) (*model.Binding, error)
	GetByID(ctx context.Context, id string) (*model.Binding, error)
	DeleteByID(ctx context.Context, id string) error
	GetListByUserID(id string, query PageQuery) (PageResult[*model.Binding], error)
}
