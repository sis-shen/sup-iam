package repository

import "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"

type UserRepository interface {
	Create(*model.User) (*model.User, error)
	GetByID(string) (*model.User, error)
	GetByUsername(string) (*model.User, error)
	GetByEmail(string) (*model.User, error)
	GetByPhone(string) (*model.User, error)
	Update(*model.User) (*model.User, error)
	DeleteByID(string) error
	GetListPage(query PageQuery) (PageResult[*model.User], error)
}
