package service

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
)

type BindingInterface interface {
	GetBindingListByUserID(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Binding], error)
	DeleteBinding(ctx context.Context, id string) error
	GetBindingById(ctx context.Context, id string) (*model.Binding, error)
	CreateBinding(ctx context.Context, binding *model.Binding) (*model.Binding, error)
}

type BindingCase struct {
	repo repository.BindingRepository
}

var _ BindingInterface = (*BindingCase)(nil)

func (bc *BindingCase) GetBindingListByUserID(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Binding], error) {
	return bc.repo.GetListByUserID(ctx, id, query)
}

func (bc *BindingCase) GetBindingById(ctx context.Context, id string) (*model.Binding, error) {
	return bc.repo.GetByID(ctx, id)
}

func (bc *BindingCase) CreateBinding(ctx context.Context, binding *model.Binding) (*model.Binding, error) {
	return bc.repo.Create(ctx, binding)
}

func (bc *BindingCase) DeleteBinding(ctx context.Context, id string) error {
	return bc.repo.DeleteByID(ctx, id)
}
