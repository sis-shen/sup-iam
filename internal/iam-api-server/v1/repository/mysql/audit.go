package mysql

import (
	"fmt"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"gorm.io/gorm"
	"strconv"
)

type AuditStore struct {
	db *gorm.DB
}

var _ repository.AuditRepository = &AuditStore{}

func (as *AuditStore) GetPolicyAuditByID(id string) (*model.PolicyAudit, error) {
	policy := &model.PolicyAudit{}
	if err := as.db.Where("id = ?", id).First(policy).Error; err != nil {
		return nil, repoError(err)
	}
	return policy, nil
}

func (as *AuditStore) GetPolicyAuditList(query repository.PageQuery) (repository.PageResult[*model.PolicyAudit], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.PolicyAudit]{}, err
	}

	db := as.db.Model(&model.PolicyAudit{})

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

	var policies []*model.PolicyAudit
	if err := db.Limit(limit).Find(&policies).Error; err != nil {
		return repository.PageResult[*model.PolicyAudit]{}, repoError(err)
	}

	result := repository.PageResult[*model.PolicyAudit]{}

	if len(policies) > mysqlQuery.Limit {
		last := policies[mysqlQuery.Limit]
		result.Items = policies[:mysqlQuery.Limit]
		result.Next = strconv.FormatUint(uint64(last.ID), 10)
	} else {
		result.Items = policies
		result.Next = ""
	}

	var total int64
	if err := db.Count(&total); err == nil {
		result.Total = total
	}

	return result, nil
}

func (as *AuditStore) GetBindingAuditByID(id string) (*model.BindingAudit, error) {
	binding := &model.BindingAudit{}
	if err := as.db.Where("id = ?", id).First(binding).Error; err != nil {
		return nil, repoError(err)
	}
	return binding, nil
}

func (as *AuditStore) GetBindingAuditList(query repository.PageQuery) (repository.PageResult[*model.BindingAudit], error) {
	mysqlQuery, err := handleQuery(&query)
	if err != nil || mysqlQuery == nil {
		return repository.PageResult[*model.BindingAudit]{}, err
	}

	db := as.db.Model(&model.BindingAudit{})

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

	var bindings []*model.BindingAudit
	if err := db.Limit(limit).Find(&bindings).Error; err != nil {
		return repository.PageResult[*model.BindingAudit]{}, repoError(err)
	}

	result := repository.PageResult[*model.BindingAudit]{}

	if len(bindings) > mysqlQuery.Limit {
		last := bindings[mysqlQuery.Limit]
		result.Items = bindings[:mysqlQuery.Limit]
		result.Next = strconv.FormatUint(uint64(last.ID), 10)
	} else {
		result.Items = bindings
		result.Next = ""
	}

	var total int64
	if err := db.Count(&total); err == nil {
		result.Total = total
	}

	return result, nil
}
