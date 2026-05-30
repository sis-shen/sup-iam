package mysql

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBindingStore_Create_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewBindingStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `bindings`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	binding := &model.Binding{Username: "testuser"}
	result, err := store.Create(context.Background(), binding)
	require.NoError(t, err)
	require.Equal(t, "testuser", result.Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBindingStore_Create_Duplicate(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewBindingStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `bindings`")).
		WillReturnError(&mysql.MySQLError{Number: 1062})
	mock.ExpectRollback()

	_, err := store.Create(context.Background(), &model.Binding{})
	require.Equal(t, repository.ErrAlreadyExists, err)
}

func TestBindingStore_GetByID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewBindingStore(gormDB)

	// BindingStore.GetByID uses Where("id = ?", id).First(binding, id) which generates:
	// SELECT * FROM `bindings` WHERE id = ? AND `bindings`.`id` = ? ORDER BY `bindings`.`id` LIMIT ?
	rows := sqlmock.NewRows([]string{"id", "secretID", "policyID", "username"}).
		AddRow(1, 100, 200, "testuser")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `bindings` WHERE id = ? AND `bindings`.`id` = ? ORDER BY `bindings`.`id` LIMIT ?")).
		WithArgs("1", "1", 1).
		WillReturnRows(rows)

	binding, err := store.GetByID(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, uint64(1), binding.ID)
	require.Equal(t, uint64(100), binding.SecretID)
	require.Equal(t, uint64(200), binding.PolicyID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBindingStore_GetByID_NotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewBindingStore(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `bindings` WHERE id = ? AND `bindings`.`id` = ? ORDER BY `bindings`.`id` LIMIT ?")).
		WithArgs("999", "999", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := store.GetByID(context.Background(), "999")
	require.Equal(t, repository.ErrNotFound, err)
}

func TestBindingStore_DeleteByID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewBindingStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `bindings` WHERE `bindings`.`id` = ?")).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.DeleteByID(context.Background(), "1")
	require.NoError(t, err)
}

func TestBindingStore_DeleteByID_NotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewBindingStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `bindings` WHERE `bindings`.`id` = ?")).
		WithArgs("999").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err := store.DeleteByID(context.Background(), "999")
	require.Equal(t, repository.ErrNotFound, err)
}

func TestBindingStore_GetListByUserID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewBindingStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(1, "user").
		AddRow(2, "user")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT `policies`.`id`,`policies`.`secretID`,`policies`.`policyID`,`policies`.`username`,`policies`.`extendShadow`,`policies`.`createdAt` FROM `policies` WHERE userID = ? ORDER BY id asc LIMIT ?")).
		WithArgs("1", 11).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `policies` WHERE userID = ?")).
		WithArgs("1").
		WillReturnRows(countRows)

	result, err := store.GetListByUserID(context.Background(), "1", repository.PageQuery{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, 2, len(result.Items))
}

func TestBindingStore_GetListByUserID_InvalidQuery(t *testing.T) {
	gormDB, _ := newMockDB(t)
	store := NewBindingStore(gormDB)

	_, err := store.GetListByUserID(context.Background(), "1", repository.PageQuery{Limit: -1})
	require.Equal(t, repository.ErrInvalidInput, err)
}
