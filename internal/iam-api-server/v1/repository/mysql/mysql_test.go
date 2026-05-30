package mysql

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/stretchr/testify/require"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Test_RepoError_Nil(t *testing.T) {
	err := repoError(nil)
	require.NoError(t, err)
}

func Test_RepoError_RecordNotFound(t *testing.T) {
	err := repoError(gorm.ErrRecordNotFound)
	require.Equal(t, repository.ErrNotFound, err)
}

func Test_RepoError_DuplicateEntry_1062(t *testing.T) {
	mysqlErr := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'foo' for key 'idx'"}
	err := repoError(mysqlErr)
	require.Equal(t, repository.ErrAlreadyExists, err)
}

func Test_RepoError_ForeignKey_1451(t *testing.T) {
	mysqlErr := &mysql.MySQLError{Number: 1451, Message: "Cannot delete or update a parent row"}
	err := repoError(mysqlErr)
	require.Equal(t, repository.ErrConflict, err)
}

func Test_RepoError_ForeignKey_1452(t *testing.T) {
	mysqlErr := &mysql.MySQLError{Number: 1452, Message: "Cannot add or update a child row"}
	err := repoError(mysqlErr)
	require.Equal(t, repository.ErrConflict, err)
}

func Test_RepoError_UnknownMySQLError(t *testing.T) {
	mysqlErr := &mysql.MySQLError{Number: 9999, Message: "unknown error"}
	err := repoError(mysqlErr)
	require.Equal(t, repository.ErrStorageFailure, err)
}

func Test_RepoError_DuplicateStringCheck(t *testing.T) {
	err := repoError(errors.New("UNIQUE constraint failed: duplicate key"))
	require.Equal(t, repository.ErrAlreadyExists, err)
}

func Test_RepoError_GenericError(t *testing.T) {
	err := repoError(errors.New("connection refused"))
	require.Equal(t, repository.ErrStorageFailure, err)
}

func Test_HandleQuery_NilQuery(t *testing.T) {
	_, err := handleQuery(nil)
	require.Equal(t, repository.ErrInvalidInput, err)
}

func Test_HandleQuery_NegativeLimit(t *testing.T) {
	_, err := handleQuery(&repository.PageQuery{Limit: -1})
	require.Equal(t, repository.ErrInvalidInput, err)
}

func Test_HandleQuery_OverLimit(t *testing.T) {
	_, err := handleQuery(&repository.PageQuery{Limit: 101})
	require.Equal(t, repository.ErrInvalidInput, err)
}

func Test_HandleQuery_InvalidCursor(t *testing.T) {
	_, err := handleQuery(&repository.PageQuery{Limit: 10, Cursor: "abc"})
	require.Equal(t, repository.ErrInvalidInput, err)
}

func Test_HandleQuery_OrderByNotAllowed(t *testing.T) {
	_, err := handleQuery(&repository.PageQuery{
		Limit:   10,
		OrderBy: "password",
	})
	require.Equal(t, repository.ErrInvalidInput, err)
}

func Test_HandleQuery_AllowlistColumns(t *testing.T) {
	allowed := []string{"id", "created_at", "updated_at", "instanceID", "username", "nickname", "name", "email", "phone"}
	for _, col := range allowed {
		q, err := handleQuery(&repository.PageQuery{Limit: 10, OrderBy: col})
		require.NoError(t, err)
		require.Equal(t, col, q.OrderBy)
	}
}

func Test_HandleQuery_Success(t *testing.T) {
	q, err := handleQuery(&repository.PageQuery{
		Limit:   10,
		Cursor:  "5",
		OrderBy: "id",
		Order:   repository.OrderDesc,
	})
	require.NoError(t, err)
	require.Equal(t, 10, q.Limit)
	require.Equal(t, uint64(5), q.Cursor)
	require.Equal(t, "id", q.OrderBy)
	require.Equal(t, repository.OrderDesc, q.Order)
}

func Test_HandleQuery_DefaultOrder(t *testing.T) {
	q, err := handleQuery(&repository.PageQuery{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, repository.OrderAsc, q.Order)
	require.Equal(t, "id", q.OrderBy)
	require.Equal(t, 0, int(q.Cursor))
}

func Test_HandleQuery_EmptyCursor(t *testing.T) {
	q, err := handleQuery(&repository.PageQuery{Limit: 10, Cursor: ""})
	require.NoError(t, err)
	require.Equal(t, uint64(0), q.Cursor)
}

// newMockDB 创建一个 gorm.DB 连接 sqlmock 的辅助函数
func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	gormDB, err := gorm.Open(gmysql.New(gmysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)
	return gormDB, mock
}
