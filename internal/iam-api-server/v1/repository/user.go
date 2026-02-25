package repository

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	DeleteByID(ctx context.Context, id string) error
	GetList(ctx context.Context, query PageQuery) (PageResult[*model.User], error)
}
