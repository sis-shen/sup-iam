package testutil

import (
	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/pkg/jwt"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"net/http"
	"net/http/httptest"
	"time"
)

func NewTestLogger() log.Logger {
	opts := log.NewOptions()
	opts.Level = log.DebugLevel.String()
	opts.EnableColor = true
	return log.New(opts)
}

func NewTestGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	return c
}

func NewTestJWTManager() jwt.Manager {
	return *jwt.NewManager(&jwt.Config{
		SecretKey:         "test-secret-key-for-unit-test",
		TokenExpireTime:   time.Hour,
		RefreshExpireTime: 24 * time.Hour,
		Issuer:            "test-issuer",
	})
}
