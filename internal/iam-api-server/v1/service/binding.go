package service

import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
)

type BindingCaseInterface interface {
	GetBindingListByUserID(ctx context.Context, id string, query repository.PageQuery) (repository.PageResult[*model.Binding], error)
	DeleteBinding(ctx context.Context, id string) error
	GetBindingById(ctx context.Context, id string) (*model.Binding, error)
	CreateBinding(ctx context.Context, binding *model.Binding) (*model.Binding, error)
}

type BindingCase struct {
	repo repository.BindingRepository
}

func NewBindingCase(repo repository.BindingRepository) *BindingCase {
	return &BindingCase{repo: repo}
}

var _ BindingCaseInterface = (*BindingCase)(nil)

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
