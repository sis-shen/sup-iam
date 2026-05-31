package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestJWTManager() *jwt.Manager {
	return jwt.NewManager(&jwt.Config{
		SecretKey:         "test-secret-for-middleware",
		TokenExpireTime:   time.Hour,
		RefreshExpireTime: 24 * time.Hour,
		Issuer:            "test",
	})
}

func generateTestToken(t *testing.T, jm *jwt.Manager) string {
	t.Helper()
	ak, _, err := jm.GenerateToken("user-001", "alice")
	require.NoError(t, err)
	return ak
}

// ---------------------------------------------------------------------------
// JWTAuthMiddleware 测试
// ---------------------------------------------------------------------------

func TestJWTAuthMiddleware_SkipPath(t *testing.T) {
	jm := newTestJWTManager()
	handler := JWTAuthMiddleware(jm, []string{"/api/v1/auth"})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)

	handler(c)
	// shouldSkip 命中，应无 abort
	assert.False(t, c.IsAborted())
}

func TestJWTAuthMiddleware_MissingToken(t *testing.T) {
	jm := newTestJWTManager()
	handler := JWTAuthMiddleware(jm, []string{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)

	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authentication token")
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	jm := newTestJWTManager()
	handler := JWTAuthMiddleware(jm, []string{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token-value")

	handler(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestJWTAuthMiddleware_Success(t *testing.T) {
	jm := newTestJWTManager()
	handler := JWTAuthMiddleware(jm, []string{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)
	c.Request.Header.Set("Authorization", "Bearer "+generateTestToken(t, jm))

	// 模拟后续 handler 来捕获 userID
	var capturedUserID string
	nextHandler := func(c *gin.Context) {
		if v, exists := c.Get(jm.GetUserIDKey()); exists {
			capturedUserID = v.(string)
		}
		c.Status(http.StatusOK)
	}

	// 先跑中间件，再跑后续 handler
	handler(c)
	nextHandler(c)

	assert.Equal(t, "user-001", capturedUserID)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// OptionalJWTAuthMiddleware 测试
// ---------------------------------------------------------------------------

func TestOptionalJWTAuthMiddleware_NoToken(t *testing.T) {
	jm := newTestJWTManager()
	handler := OptionalJWTAuthMiddleware(jm)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public", nil)

	nextCalled := false
	nextHandler := func(c *gin.Context) {
		nextCalled = true
		_, exists := c.Get(jm.GetUserIDKey())
		assert.False(t, exists, "userID should not be set when no token")
		c.Status(http.StatusOK)
	}

	handler(c)
	nextHandler(c)

	assert.True(t, nextCalled)
}

func TestOptionalJWTAuthMiddleware_ValidToken(t *testing.T) {
	jm := newTestJWTManager()
	handler := OptionalJWTAuthMiddleware(jm)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public", nil)
	c.Request.Header.Set("Authorization", "Bearer "+generateTestToken(t, jm))

	var capturedUserID string
	nextHandler := func(c *gin.Context) {
		if v, exists := c.Get(jm.GetUserIDKey()); exists {
			capturedUserID = v.(string)
		}
		c.Status(http.StatusOK)
	}

	handler(c)
	nextHandler(c)

	assert.Equal(t, "user-001", capturedUserID)
}

func TestOptionalJWTAuthMiddleware_InvalidToken(t *testing.T) {
	jm := newTestJWTManager()
	handler := OptionalJWTAuthMiddleware(jm)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/public", nil)
	c.Request.Header.Set("Authorization", "Bearer garbage-token")

	nextCalled := false
	nextHandler := func(c *gin.Context) {
		nextCalled = true
		_, exists := c.Get(jm.GetUserIDKey())
		assert.False(t, exists, "userID should not be set for invalid token")
		c.Status(http.StatusOK)
	}

	handler(c)
	nextHandler(c)
	assert.True(t, nextCalled)
}

// ---------------------------------------------------------------------------
// shouldSkip 测试
// ---------------------------------------------------------------------------

func TestShouldSkip_Matching(t *testing.T) {
	assert.True(t, shouldSkip("/api/v1/auth/login", []string{"/api/v1/auth"}))
	assert.True(t, shouldSkip("/api/v1/auth/register", []string{"/api/v1/auth"}))
	assert.True(t, shouldSkip("/api/v1/auth", []string{"/api/v1/auth"}))
}

func TestShouldSkip_NotMatching(t *testing.T) {
	assert.False(t, shouldSkip("/api/v1/secrets", []string{"/api/v1/auth"}))
}

func TestShouldSkip_EmptySkipList(t *testing.T) {
	assert.False(t, shouldSkip("/api/v1/auth/login", []string{}))
}

func TestShouldSkip_MultipleSkipPaths(t *testing.T) {
	assert.True(t, shouldSkip("/api/v1/auth/login", []string{"/healthz", "/metrics", "/api/v1/auth"}))
	assert.True(t, shouldSkip("/healthz", []string{"/healthz", "/metrics", "/api/v1/auth"}))
	assert.False(t, shouldSkip("/api/v1/secrets", []string{"/healthz", "/metrics", "/api/v1/auth"}))
}
