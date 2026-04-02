package mysql

import (
	"context"
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"gorm.io/gorm"
	"strconv"
)

type BindingStore struct {
	db *gorm.DB
}

func NewBindingStore(db *gorm.DB) *BindingStore {
	return &BindingStore{db: db}
}

var _ repository.BindingRepository = &BindingStore{}

func (bs *BindingStore) Create(ctx context.Context, binding *model.Binding) (*model.Binding, error) {
	if err := bs.db.WithContext(ctx).Create(binding).Error; err != nil {
		return nil, repoError(err)
	}
	return binding, nil
}

func (bs *BindingStore) GetByID(ctx context.Context, id string) (*model.Binding, error) {
	var binding = &model.Binding{}
	if err := bs.db.WithContext(ctx).Where("id = ?", id).First(binding, id).Error; err != nil {
		return nil, repoError(err)
	}
	return binding, nil
}

func (bs *BindingStore) DeleteByID(ctx context.Context, id string) error {
	if err := bs.db.WithContext(ctx).Delete(&model.Binding{}, id).Error; err != nil {
		return repoError(err)
	}
	return nil
}

func (bs *BindingStore) GetListByUserID(ctx context.Context, userID string, query repository.PageQuery) (repository.PageResult[*model.Binding], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.Binding]{}, err
	}

	db := bs.db.WithContext(ctx).Model(&model.Policy{})

	db = db.Where("userID = ?", userID)

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

	var policies []*model.Binding
	if err := db.Limit(limit).Find(&policies).Error; err != nil {
		return repository.PageResult[*model.Binding]{}, repoError(err)
	}

	result := repository.PageResult[*model.Binding]{}
	if len(policies) > mysqlQuery.Limit {
		last := policies[mysqlQuery.Limit-1]
		result.Items = policies[:mysqlQuery.Limit]
		result.Next = strconv.FormatUint(uint64(last.ID), 10)
	} else {
		result.Items = policies
		result.Next = ""
	}

	var total int64
	if err := bs.db.WithContext(ctx).
		Model(&model.Policy{}).
		Where("userID = ?", userID).
		Count(&total).Error; err != nil {
		result.Total = total
	}
	return result, nil
}
