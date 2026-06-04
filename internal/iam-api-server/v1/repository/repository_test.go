package repository

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrorSentinels(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{name: "ErrNotFound", err: ErrNotFound, msg: "repository: not found"},
		{name: "ErrAlreadyExists", err: ErrAlreadyExists, msg: "repository: already exists"},
		{name: "ErrConflict", err: ErrConflict, msg: "repository: conflict"},
		{name: "ErrInvalidInput", err: ErrInvalidInput, msg: "repository: invalid input"},
		{name: "ErrStorageFailure", err: ErrStorageFailure, msg: "repository: storage failure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.msg, tt.err.Error())
			// Verify they can be properly detected with errors.Is
			assert.True(t, errors.Is(tt.err, tt.err), "errors.Is should detect self")
			// Verify they're not nil
			assert.NotNil(t, tt.err)
		})
	}
}

func TestErrorSentinels_Uniqueness(t *testing.T) {
	// All error sentinels should be distinct
	errs := []error{ErrNotFound, ErrAlreadyExists, ErrConflict, ErrInvalidInput, ErrStorageFailure}
	for i, ei := range errs {
		for j, ej := range errs {
			if i != j {
				assert.False(t, errors.Is(ei, ej), "%s should not match %s", ei.Error(), ej.Error())
			}
		}
	}
}

func TestErrorSentinels_Wrapping(t *testing.T) {
	// Test that wrapping preserves error identity
	wrapped := errors.Join(ErrNotFound, errors.New("additional context"))
	assert.True(t, errors.Is(wrapped, ErrNotFound), "wrapped error should preserve ErrNotFound")
}

func TestOrderConstants(t *testing.T) {
	assert.Equal(t, Order("asc"), OrderAsc)
	assert.Equal(t, Order("desc"), OrderDesc)
	assert.NotEqual(t, OrderAsc, OrderDesc)
}

func TestOrder_ValidValues(t *testing.T) {
	assert.Equal(t, "asc", string(OrderAsc))
	assert.Equal(t, "desc", string(OrderDesc))
}

func TestPageQuery_DefaultValues(t *testing.T) {
	q := PageQuery{}
	assert.Equal(t, 0, q.Limit)
	assert.Equal(t, "", q.Cursor)
	assert.Equal(t, "", q.OrderBy)
	assert.Equal(t, Order(""), q.Order)
}

func TestPageQuery_WithValues(t *testing.T) {
	q := PageQuery{
		Limit:   20,
		Cursor:  "100",
		OrderBy: "createdAt",
		Order:   OrderDesc,
	}
	assert.Equal(t, 20, q.Limit)
	assert.Equal(t, "100", q.Cursor)
	assert.Equal(t, "createdAt", q.OrderBy)
	assert.Equal(t, OrderDesc, q.Order)
}

func TestPageResult_ZeroValues(t *testing.T) {
	r := PageResult[*struct{}]{}
	assert.Nil(t, r.Items)
	assert.Equal(t, int64(0), r.Total)
	assert.Equal(t, "", r.Next)
}

func TestPageResult_EmptyItems(t *testing.T) {
	r := PageResult[*struct{}]{
		Items: []*struct{}{},
		Total: 0,
		Next:  "",
	}
	assert.Empty(t, r.Items)
	assert.Equal(t, int64(0), r.Total)
}

func TestPageResult_WithItems(t *testing.T) {
	type item struct {
		ID string
	}
	i1 := &item{ID: "1"}
	i2 := &item{ID: "2"}
	r := PageResult[*item]{
		Items: []*item{i1, i2},
		Total: 2,
		Next:  "cursor-next",
	}
	assert.Len(t, r.Items, 2)
	assert.Equal(t, "1", r.Items[0].ID)
	assert.Equal(t, int64(2), r.Total)
	assert.Equal(t, "cursor-next", r.Next)
}
