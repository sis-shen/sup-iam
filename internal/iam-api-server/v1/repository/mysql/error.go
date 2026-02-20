package mysql

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
)

// repoError 将底层存储错误（gorm / mysql）
// 映射为 repository 层稳定错误语义
func repoError(err error) error {
	if err == nil {
		return nil
	}

	// gorm 层通用错误
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return repository.ErrNotFound
	}

	// MySQL driver 错误（需要引入 go-sql-driver/mysql）
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			// Duplicate entry (唯一索引冲突)
			return repository.ErrAlreadyExists
		case 1451, 1452:
			// 外键约束失败
			return repository.ErrConflict
		default:
			return repository.ErrStorageFailure
		}
	}

	// 兜底：部分 GORM 错误是字符串包装的
	if strings.Contains(err.Error(), "duplicate") {
		return repository.ErrAlreadyExists
	}

	return repository.ErrStorageFailure
}
