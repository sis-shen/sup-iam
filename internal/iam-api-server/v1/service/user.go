package service

//go:generate mockgen -destination=./mock/user_case_mock.go -package=mock . UserCaseInterface
import (
	"context"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
)

type UserCaseInterface interface {
	GetUserList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.User], error)
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*model.User, error)
	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) (*model.User, error)
	DeleteUser(ctx context.Context, id string) error
	HashPassword(plainPassword string) (string, error)
	VerifyPassword(plainPassword string, hashPassword string) error
}
type UserCase struct {
	repo   repository.UserRepository
	hasher PasswordHasherInterface
}

// 确保 UserCase 实现了 UserCaseInterface
var _ UserCaseInterface = (*UserCase)(nil)

func NewUserCase(repo repository.UserRepository) *UserCase {
	return &UserCase{repo: repo}
}

func (uc *UserCase) GetUserList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.User], error) {
	return uc.repo.GetList(ctx, query)
}

func (uc *UserCase) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *UserCase) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return uc.repo.GetByEmail(ctx, email)
}

func (uc *UserCase) GetUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	return uc.repo.GetByPhone(ctx, phone)
}

func (uc *UserCase) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return uc.repo.GetByUsername(ctx, username)
}

func (uc *UserCase) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	return uc.repo.Create(ctx, user)
}

func (uc *UserCase) UpdateUser(ctx context.Context, user *model.User) (*model.User, error) {
	return uc.repo.Update(ctx, user)
}

func (uc *UserCase) DeleteUser(ctx context.Context, id string) error {
	return uc.repo.DeleteByID(ctx, id)
}

func (uc *UserCase) HashPassword(plainPassword string) (string, error) {
	return uc.hasher.HashPassword(plainPassword)
}

func (uc *UserCase) VerifyPassword(plainPassword string, hashPassword string) error {
	return uc.hasher.VerifyPassword(plainPassword, hashPassword)
}
