package repository

import "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"

type BindingRepository interface {
	Create(*model.Binding) (*model.Binding, error)
	GetByID(string) (*model.Binding, error)
	Delete(string) error
	GetListByUserID(userID int, query PageQuery) (PageResult[*model.Binding], error)
}
