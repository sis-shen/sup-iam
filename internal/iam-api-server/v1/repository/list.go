package repository

type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

type PageQuery struct {
	Limit   int
	Cursor  string
	OrderBy string
	Order   Order
}

type PageResult[T any] struct {
	Items []T
	Total int64
	Next  string // 下一页 cursor
}
