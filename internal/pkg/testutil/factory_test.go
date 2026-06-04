package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTestLogger(t *testing.T) {
	logger := NewTestLogger()
	assert.NotNil(t, logger)
	assert.True(t, logger.Enabled())
}

func TestNewTestGinContext(t *testing.T) {
	ctx := NewTestGinContext()
	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.Request)
	assert.Equal(t, "/api/v1/auth/login", ctx.Request.URL.Path)
	assert.Equal(t, "POST", ctx.Request.Method)
}

func TestNewTestJWTManager(t *testing.T) {
	mgr := NewTestJWTManager()
	assert.NotNil(t, mgr)
	assert.NotNil(t, mgr)
}
