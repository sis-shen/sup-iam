package mysql

import (
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"gorm.io/gorm"
	"strconv"
)

type PolicyStore struct {
	db *gorm.DB
}

var _ repository.PolicyRepository = &PolicyStore{}

func (ps *PolicyStore) Create(policy *model.Policy) (*model.Policy, error) {
	if err := ps.db.Create(policy).Error; err != nil {
		return nil, repoError(err)
	}
	return policy, nil
}

func (ps *PolicyStore) GetByID(policyID string) (*model.Policy, error) {
	policy := &model.Policy{}
	if err := ps.db.Where("id = ?", policyID).First(policy).Error; err != nil {
		return nil, repoError(err)
	}
	return policy, nil
}

func (ps *PolicyStore) Update(policy *model.Policy) (*model.Policy, error) {
	if err := ps.db.Save(policy).Error; err != nil {
		return nil, repoError(err)
	}
	return policy, nil
}

func (ps *PolicyStore) DeleteByID(policyID string) error {
	if err := ps.db.Delete(&model.Policy{}, "id = ?", policyID).Error; err != nil {
		return repoError(err)
	}
	return nil
}

func (ps *PolicyStore) GetListByUserID(userID string, query repository.PageQuery) (repository.PageResult[*model.Policy], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.Policy]{}, err
	}

	db := ps.db.Model(&model.Policy{})

	db = db.Where("userID = ?", userID)

	//游标条件
	if mysqlQuery.Cursor > 0 {
		if mysqlQuery.Order == repository.OrderAsc {
			db = db.Where("id > ?", mysqlQuery.Cursor)
		} else {
			db = db.Where("id < ?", mysqlQuery.Cursor)
		}
	}
	// 排序
	orderClause := fmt.Sprintf("%s %s", mysqlQuery.OrderBy, mysqlQuery.Order)
	db = db.Order(orderClause)
	//查一下有没有下一页
	limit := mysqlQuery.Limit + 1

	var policies []*model.Policy
	if err := db.Limit(limit).Find(&policies).Error; err != nil {
		return repository.PageResult[*model.Policy]{}, repoError(err)
	}

	result := repository.PageResult[*model.Policy]{}
	if len(policies) > mysqlQuery.Limit {
		last := policies[mysqlQuery.Limit-1]
		result.Items = policies[:mysqlQuery.Limit]
		result.Next = strconv.FormatUint(uint64(last.ID), 10)
	} else {
		result.Items = policies
		result.Next = ""
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		result.Total = total
	}
	return result, nil
}

func (ps *PolicyStore) GetSecretListByPolicyID(policyID string, query repository.PageQuery) (repository.PageResult[*model.Secret], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.Secret]{}, err
	}
	db := ps.db.Model(&model.Secret{})
	db = db.Joins(`
    JOIN secret_policy_binding spb
      ON spb.policyID = secrets.id
	`)
	db = db.Where("policyID = ?", policyID)

	//游标条件
	if mysqlQuery.Cursor > 0 {
		if mysqlQuery.Order == repository.OrderAsc {
			db = db.Where("id > ?", mysqlQuery.Cursor)
		} else {
			db = db.Where("id < ?", mysqlQuery.Cursor)
		}
	}
	// 排序
	orderClause := fmt.Sprintf("%s %s", mysqlQuery.OrderBy, mysqlQuery.Order)
	db = db.Order(orderClause)
	//查一下有没有下一页
	limit := mysqlQuery.Limit + 1

	var policies []*model.Secret
	if err := db.Limit(limit).Find(&policies).Error; err != nil {
		return repository.PageResult[*model.Secret]{}, repoError(err)
	}
	result := repository.PageResult[*model.Secret]{}
	if len(policies) > mysqlQuery.Limit {
		last := policies[mysqlQuery.Limit]
		result.Items = policies[:mysqlQuery.Limit]
		result.Next = strconv.FormatUint(uint64(last.ID), 10)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		result.Total = total
	}
	return result, nil
}
