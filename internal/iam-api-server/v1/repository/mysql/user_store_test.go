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

func TestUserStore_Create_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &model.User{Username: "testuser"}
	result, err := store.Create(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, "testuser", result.Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserStore_Create_Duplicate(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `users`")).
		WillReturnError(&mysql.MySQLError{Number: 1062})
	mock.ExpectRollback()

	_, err := store.Create(context.Background(), &model.User{})
	require.Equal(t, repository.ErrAlreadyExists, err)
}

func TestUserStore_GetByID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "username", "nickname"}).
		AddRow(1, "testuser", "Test")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("1", 1).
		WillReturnRows(rows)

	user, err := store.GetByID(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, uint64(1), user.ID)
	require.Equal(t, "testuser", user.Username)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUserStore_GetByID_NotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("999", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	_, err := store.GetByID(context.Background(), "999")
	require.Equal(t, repository.ErrNotFound, err)
}

func TestUserStore_GetByUsername_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(1, "testuser")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE username = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("testuser", 1).
		WillReturnRows(rows)

	user, err := store.GetByUsername(context.Background(), "testuser")
	require.NoError(t, err)
	require.Equal(t, uint64(1), user.ID)
	require.Equal(t, "testuser", user.Username)
}

func TestUserStore_GetByEmail_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "email"}).
		AddRow(1, "test@example.com")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE email = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("test@example.com", 1).
		WillReturnRows(rows)

	user, err := store.GetByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestUserStore_GetByPhone_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "phone"}).
		AddRow(1, "13800138000")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE phone = ? ORDER BY `users`.`id` LIMIT ?")).
		WithArgs("13800138000", 1).
		WillReturnRows(rows)

	user, err := store.GetByPhone(context.Background(), "13800138000")
	require.NoError(t, err)
	require.NotNil(t, user)
}

func TestUserStore_Update_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `users` SET")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	user := &model.User{ID: 1, Username: "updated"}
	result, err := store.Update(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, uint64(1), result.ID)
}

func TestUserStore_DeleteByID_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `users` WHERE id = ?")).
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := store.DeleteByID(context.Background(), "1")
	require.NoError(t, err)
}

func TestUserStore_DeleteByID_NotFound(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM `users` WHERE id = ?")).
		WithArgs("999").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectRollback()

	err := store.DeleteByID(context.Background(), "999")
	require.Equal(t, repository.ErrNotFound, err)
}

func TestUserStore_ExistsByUsername_Exists(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE username = ?")).
		WithArgs("testuser").
		WillReturnRows(rows)

	exists, err := store.ExistsByUsername(context.Background(), "testuser")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestUserStore_ExistsByUsername_NotExists(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE username = ?")).
		WithArgs("nonexistent").
		WillReturnRows(rows)

	exists, err := store.ExistsByUsername(context.Background(), "nonexistent")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestUserStore_ExistsByEmail(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE email = ?")).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	exists, err := store.ExistsByEmail(context.Background(), "test@example.com")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestUserStore_ExistsByPhone(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users` WHERE phone = ?")).
		WithArgs("13800138000").
		WillReturnRows(rows)

	exists, err := store.ExistsByPhone(context.Background(), "13800138000")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestUserStore_GetList_Success(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(1, "alice").
		AddRow(2, "bob")

	// cursor = 0 时没有 WHERE 条件
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` ORDER BY id asc LIMIT ?")).
		WithArgs(11).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(10)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users`")).
		WillReturnRows(countRows)

	result, err := store.GetList(context.Background(), repository.PageQuery{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, 2, len(result.Items))
	require.Equal(t, int64(10), result.Total)
	require.Equal(t, "", result.Next)
}

func TestUserStore_GetList_WithCursor(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	rows := sqlmock.NewRows([]string{"id", "username"}).
		AddRow(5, "eve").
		AddRow(6, "frank").
		AddRow(7, "grace").
		AddRow(8, "heidi").
		AddRow(9, "ivan").
		AddRow(10, "judy")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` WHERE id >= ? ORDER BY id asc LIMIT ?")).
		WithArgs(5, 6).
		WillReturnRows(rows)

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(20)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `users`")).
		WillReturnRows(countRows)

	result, err := store.GetList(context.Background(), repository.PageQuery{
		Limit: 5, Cursor: "5",
	})
	require.NoError(t, err)
	require.Equal(t, 5, len(result.Items))
	require.Equal(t, "10", result.Next)
}

func TestUserStore_GetList_InvalidQuery(t *testing.T) {
	gormDB, _ := newMockDB(t)
	store := NewUserStore(gormDB)

	_, err := store.GetList(context.Background(), repository.PageQuery{Limit: -1})
	require.Equal(t, repository.ErrInvalidInput, err)
}

func TestUserStore_GetList_DBError(t *testing.T) {
	gormDB, mock := newMockDB(t)
	store := NewUserStore(gormDB)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `users` ORDER BY")).
		WillReturnError(gorm.ErrInvalidDB)

	_, err := store.GetList(context.Background(), repository.PageQuery{Limit: 10})
	require.Error(t, err)
}
