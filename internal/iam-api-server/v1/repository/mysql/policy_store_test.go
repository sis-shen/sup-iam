package mysql

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPolicyStore_Create_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `policies`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	policy := &model.Policy{Name: "test-policy"}
	result, err := store.Create(context.Background(), policy)
	require.NoError(t, err)
	require.Equal(t, "test-policy", result.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPolicyStore_Create_Duplicate(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `policies`")).
		WillReturnError(&mysql.MySQLError{Number: 1062})
	mock.ExpectRollback()

	_, err := store.Create(context.Background(), &model.Policy{})
	require.Equal(t, repository.ErrAlreadyExists, err)
}

func TestPolicyStore_GetByID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "name", "username", "createdAt", "updatedAt"}).
		AddRow(1, "policy-1", "user", now, now)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `policies` WHERE id = ? ORDER BY `policies`.`id` LIMIT ?")).
		WithArgs("1", 1).
		WillReturnRows(rows)

	policy, err := store.GetByID(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, uint64(1), policy.ID)
	require.Equal(t, "policy-1", policy.Name)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPolicyStore_GetByID_NotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `policies` WHERE id = ? ORDER BY `policies`.`id` LIMIT ?")).
		WithArgs("999", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := store.GetByID(context.Background(), "999")
	require.Equal(t, repository.ErrNotFound, err)
}

func TestPolicyStore_Update_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `policies` SET")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	policy := &model.Policy{ID: 1, Name: "updated-policy"}
	result, err := store.Update(context.Background(), policy)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.ID)
}

func TestPolicyStore_DeleteByID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `policies` WHERE id = ?")).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.DeleteByID(context.Background(), "1")
	require.NoError(t, err)
}

func TestPolicyStore_DeleteByID_NotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `policies` WHERE id = ?")).
		WithArgs("999").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err := store.DeleteByID(context.Background(), "999")
	require.Equal(t, repository.ErrNotFound, err)
}

func TestPolicyStore_GetListByUserID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "name", "username"}).
		AddRow(1, "policy-1", "user").
		AddRow(2, "policy-2", "user")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `policies` WHERE userID = ? ORDER BY id asc LIMIT ?")).
		WithArgs("1", 11).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `policies` WHERE userID = ? LIMIT ?")).
		WithArgs("1", 11).
		WillReturnRows(countRows)

	result, err := store.GetListByUserID(context.Background(), "1", repository.PageQuery{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, 2, len(result.Items))
}

func TestPolicyStore_GetListByUserID_InvalidQuery(t *testing.T) {
	gormDB, _ := newMockDB(t)
	store := NewPolicyStore(gormDB)

	_, err := store.GetListByUserID(context.Background(), "1", repository.PageQuery{Limit: -1})
	require.Equal(t, repository.ErrInvalidInput, err)
}

func TestPolicyStore_GetSecretListByPolicyID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewPolicyStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "accessKey"}).
		AddRow(1, "ak-001")

	mock.ExpectQuery(`JOIN secret_policy_binding spb`).
		WithArgs("1", 11).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(3)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `secrets` WHERE policyID = ?")).
		WithArgs("1").
		WillReturnRows(countRows)

	result, err := store.GetSecretListByPolicyID(context.Background(), "1", repository.PageQuery{Limit: 10})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
	_ = result
}

func TestPolicyStore_GetSecretListByPolicyID_InvalidQuery(t *testing.T) {
	gormDB, _ := newMockDB(t)
	store := NewPolicyStore(gormDB)

	_, err := store.GetSecretListByPolicyID(context.Background(), "1", repository.PageQuery{Limit: -1})
	require.Equal(t, repository.ErrInvalidInput, err)
}
