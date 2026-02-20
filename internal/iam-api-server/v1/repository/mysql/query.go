package mysql

import (
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"strconv"
)

type mySQLPageQuery struct {
	Limit   int              //SQL中的limit
	OrderBy string           //排序依据的列
	Order   repository.Order //正序还是反序
	Cursor  uint64           //指针指向的自增主键
}

func parseCursor(cursor string) (uint64, error) {
	if cursor == "" {
		return 0, nil
	}
	return strconv.ParseUint(cursor, 10, 64)
}

func parseOrderBy(orderBy string) (string, error) {
	if orderBy == "" {
		return "id", nil
	}
	col, ok := allowedOrderBy[orderBy]
	if !ok {
		return "", repository.ErrInvalidInput
	}
	return col, nil
}

func handleQuery(query *repository.PageQuery) (*mySQLPageQuery, error) {
	sqlQuery := &mySQLPageQuery{}
	if query == nil {
		return nil, repository.ErrInvalidInput
	}
	if query.Limit < 0 || query.Limit > 100 {
		return nil, repository.ErrInvalidInput
	} else {
		sqlQuery.Limit = query.Limit
	}

	cursorID, err := parseCursor(query.Cursor)
	if err != nil {
		return nil, repository.ErrInvalidInput
	}
	sqlQuery.Cursor = cursorID

	col, err := parseOrderBy(query.OrderBy)
	if err != nil {
		return nil, repository.ErrInvalidInput
	}
	sqlQuery.OrderBy = col

	if query.Order == "" {
		sqlQuery.Order = repository.OrderAsc
	} else {
		sqlQuery.Order = query.Order
	}

	return sqlQuery, nil
}
