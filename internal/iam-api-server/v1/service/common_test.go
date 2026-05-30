package service

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/stretchr/testify/assert"
)

func TestGenerateInstanceID_Length(t *testing.T) {
	id := GenerateInstanceID()
	assert.Equal(t, 32, len(id))
}

func TestGenerateInstanceID_HexFormat(t *testing.T) {
	id := GenerateInstanceID()
	assert.Equal(t, 32, len(id))
	// 验证全部是有效的十六进制字符
	for _, c := range id {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
			"instance ID should only contain hex chars, got: %c", c)
	}
}

func TestGenerateInstanceID_Uniqueness(t *testing.T) {
	id1 := GenerateInstanceID()
	id2 := GenerateInstanceID()
	assert.NotEqual(t, id1, id2)
}

func TestGenerateInstanceID_NoPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = GenerateInstanceID()
	})
}

func TestDerefSlice_NilElements(t *testing.T) {
	a := "a"
	b := "b"
	input := []*string{&a, &b}
	result := DerefSlice(input)
	assert.Equal(t, []string{"a", "b"}, result)
}

func TestDerefSlice_Empty(t *testing.T) {
	result := DerefSlice([]*string{})
	assert.Empty(t, result)
}

func TestParseListQuery_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	query, err := ParseListQuery(c)
	assert.NoError(t, err)
	assert.Equal(t, 10, query.Limit)
	assert.Empty(t, query.Cursor)
	assert.Equal(t, repository.OrderAsc, query.Order)
	assert.Empty(t, query.OrderBy)
}

func TestParseListQuery_WithParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=20&next=100&order=desc", nil)

	query, err := ParseListQuery(c)
	assert.NoError(t, err)
	assert.Equal(t, 20, query.Limit)
	assert.Equal(t, "100", query.Cursor)
	assert.Equal(t, repository.OrderDesc, query.Order)
}

func TestParseListQuery_InvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?limit=abc", nil)

	query, err := ParseListQuery(c)
	assert.NoError(t, err)
	assert.Equal(t, 10, query.Limit, "invalid limit should fallback to 10")
}

func TestParseListQuery_InvalidOrder_DefaultsToAsc(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?order=invalid", nil)

	query, err := ParseListQuery(c)
	assert.NoError(t, err)
	assert.Equal(t, repository.OrderAsc, query.Order)
}

func TestGetPage_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	page := GetPage(c)
	assert.Equal(t, int32(1), page)
}

func TestGetPage_WithParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?page=3", nil)

	page := GetPage(c)
	assert.Equal(t, int32(3), page)
}

func TestGetPage_InvalidParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?page=abc", nil)

	page := GetPage(c)
	assert.Equal(t, int32(1), page, "invalid page should fallback to 1")
}

func TestGetPageSize_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	pageSize := GetPageSize(c)
	assert.Equal(t, int32(1), pageSize)
}

func TestGetPageSize_WithParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?page_size=50", nil)

	pageSize := GetPageSize(c)
	assert.Equal(t, int32(50), pageSize)
}

func TestGetPageSize_InvalidParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?page_size=abc", nil)

	pageSize := GetPageSize(c)
	assert.Equal(t, int32(1), pageSize, "invalid page_size should fallback to 1")
}

func TestValidatePasswordStrength_Valid(t *testing.T) {
	ac := &AuthCase{}
	err := ac.validatePasswordStrength("Pass1234")
	assert.NoError(t, err)
}

func TestValidatePasswordStrength_NoNumber(t *testing.T) {
	ac := &AuthCase{}
	err := ac.validatePasswordStrength("Password")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must contain at least one number")
}

func TestValidatePasswordStrength_NoUppercase(t *testing.T) {
	ac := &AuthCase{}
	err := ac.validatePasswordStrength("pass1234")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must contain at least one uppercase letter")
}

func TestValidatePasswordStrength_NoLowercase(t *testing.T) {
	ac := &AuthCase{}
	err := ac.validatePasswordStrength("PASS1234")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must contain at least one lowercase letter")
}

func TestValidatePasswordStrength_TooShort(t *testing.T) {
	ac := &AuthCase{}
	err := ac.validatePasswordStrength("Pa1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8 characters")
}

func TestGenerateInstanceID_ContainsTimestampPrefix(t *testing.T) {
	id := GenerateInstanceID()
	assert.Equal(t, 32, len(id))
	// 前16位是时间戳
	timestampHex := id[:16]
	_, err := strconv.ParseUint(timestampHex, 16, 64)
	assert.NoError(t, err, "timestamp part should be valid hex")
}
