package mysql

import (
	"context"
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"gorm.io/gorm"
	"strconv"
)

type UserStore struct {
	db *gorm.DB
}

var _ repository.UserRepository = &UserStore{}

func (us *UserStore) Create(ctx context.Context, user *model.User) (*model.User, error) {
	if err := us.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, repoError(err)
	}
	return user, nil
}

func (us *UserStore) GetByID(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	if err := us.db.WithContext(ctx).Where("id = ?", id).First(user).Error; err != nil {
		return nil, repoError(err)
	}
	return user, nil
}

func (us *UserStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	if err := us.db.WithContext(ctx).Where("username = ?", username).First(user).Error; err != nil {
		return nil, repoError(err)
	}
	return user, nil
}

func (us *UserStore) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	if err := us.db.WithContext(ctx).Where("email = ?", email).First(user).Error; err != nil {
		return nil, repoError(err)
	}
	return user, nil
}

func (us *UserStore) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	user := &model.User{}
	if err := us.db.WithContext(ctx).Where("phone = ?", phone).First(user).Error; err != nil {
		return nil, repoError(err)
	}
	return user, nil
}
func (us *UserStore) Update(ctx context.Context, user *model.User) (*model.User, error) {
	if err := us.db.WithContext(ctx).Save(user).Error; err != nil {
		return nil, repoError(err)
	}
	return user, nil
}

func (us *UserStore) DeleteByID(ctx context.Context, id string) error {
	if err := us.db.WithContext(ctx).Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
		return repoError(err)
	}
	return nil
}

func (us *UserStore) GetList(ctx context.Context, query repository.PageQuery) (repository.PageResult[*model.User], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.User]{}, err
	}

	db := us.db.WithContext(ctx).Model(&model.User{})

	//游标条件
	if mysqlQuery.Cursor > 0 {
		if mysqlQuery.Order == repository.OrderAsc {
			db = db.Where("id >= ?", mysqlQuery.Cursor)
		} else {
			db = db.Where("id <= ?", mysqlQuery.Cursor)
		}
	}
	// 排序
	orderClause := fmt.Sprintf("%s %s", mysqlQuery.OrderBy, mysqlQuery.Order)
	db = db.Order(orderClause)
	//查一下有没有下一页
	limit := mysqlQuery.Limit + 1

	var users []*model.User
	if err := db.Limit(limit).Find(&users).Error; err != nil {
		return repository.PageResult[*model.User]{}, repoError(err)
	}

	result := repository.PageResult[*model.User]{}

	if len(users) > mysqlQuery.Limit {
		last := users[mysqlQuery.Limit]
		result.Items = users[:mysqlQuery.Limit]
		result.Next = strconv.FormatUint(
			uint64(last.ID), 10)
	} else {
		result.Items = users
		result.Next = ""
	}

	var total int64
	if err := us.db.
		WithContext(ctx).
		Model(&model.User{}).
		Count(&total).Error; err == nil {
		result.Total = total
	}

	return result, nil
}
