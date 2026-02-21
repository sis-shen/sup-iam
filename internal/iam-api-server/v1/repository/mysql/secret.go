package mysql

import (
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"gorm.io/gorm"
	"strconv"
)

type SecretStore struct {
	db *gorm.DB
}

var _ repository.SecretRepository = &SecretStore{}

func (ss *SecretStore) Create(secret *model.Secret) (*model.Secret, error) {
	if err := ss.db.Create(secret).Error; err != nil {
		return nil, repoError(err)
	}
	return secret, nil
}

func (ss *SecretStore) GetByID(id string) (*model.Secret, error) {
	secret := &model.Secret{}
	if err := ss.db.Where("id = ?", id).First(secret).Error; err != nil {
		return nil, repoError(err)
	}
	return secret, nil
}

func (ss *SecretStore) Update(secret *model.Secret) (*model.Secret, error) {
	if err := ss.db.Save(secret).Error; err != nil {
		return nil, repoError(err)
	}
	return secret, nil
}

func (ss *SecretStore) DeleteByID(id string) error {
	if err := ss.db.Delete(&model.Secret{}, "id = ?", id).Error; err != nil {
		return repoError(err)
	}
	return nil
}

func (ss *SecretStore) GetListByUserID(userID string, query repository.PageQuery) (repository.PageResult[*model.Secret], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.Secret]{}, err
	}

	db := ss.db.Model(&model.Secret{})

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

	var secrets []*model.Secret
	if err := db.Limit(limit).Find(&secrets).Error; err != nil {
		return repository.PageResult[*model.Secret]{}, repoError(err)
	}

	result := repository.PageResult[*model.Secret]{}
	if len(secrets) > mysqlQuery.Limit {
		last := secrets[mysqlQuery.Limit-1]
		result.Items = secrets[:mysqlQuery.Limit]
		result.Next = strconv.FormatUint(uint64(last.ID), 10)
	} else {
		result.Items = secrets
		result.Next = ""
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		result.Total = total
	}
	return result, nil
}

// SELECT p.*
// FROM policies p
// JOIN secret_policy_binding spb
// ON spb.policyID = p.id
// WHERE spb.secretID = ?
// ORDER BY p.id ASC
// LIMIT ?, ?
func (ss *SecretStore) GetPolicyListBySecretID(secretID string, query repository.PageQuery) (repository.PageResult[*model.Policy], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.Policy]{}, err
	}
	db := ss.db.Model(&model.Policy{})
	db = db.Joins(`
    JOIN secret_policy_binding spb
      ON spb.policyID = policies.id
	`)
	db = db.Where("secretID = ?", secretID)

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
