package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCachedSecret(t *testing.T) {
	expiresAt := time.Now().Add(24 * time.Hour)

	s := &CachedSecret{
		ID:        "secret-001",
		AccessKey: "AKID123456",
		SecretKey: "SK123456",
		ExpiredAt: expiresAt,
	}

	assert.Equal(t, "secret-001", s.ID)
	assert.Equal(t, "AKID123456", s.AccessKey)
	assert.Equal(t, "SK123456", s.SecretKey)
	assert.Equal(t, expiresAt, s.ExpiredAt)
}

func TestCachedSecret_ZeroValues(t *testing.T) {
	s := &CachedSecret{}
	assert.Equal(t, "", s.ID)
	assert.Equal(t, "", s.AccessKey)
	assert.Equal(t, "", s.SecretKey)
	assert.True(t, s.ExpiredAt.IsZero())
}

func TestCachedPolicy(t *testing.T) {
	p := &CachedPolicy{
		ID:       "policy-001",
		SecretID: "secret-001",
		Username: "testuser",
		DSL:      "p, user, resource, action",
	}

	assert.Equal(t, "policy-001", p.ID)
	assert.Equal(t, "secret-001", p.SecretID)
	assert.Equal(t, "testuser", p.Username)
	assert.Equal(t, "p, user, resource, action", p.DSL)
}

func TestCachedPolicy_ZeroValues(t *testing.T) {
	p := &CachedPolicy{}
	assert.Equal(t, "", p.ID)
	assert.Equal(t, "", p.SecretID)
	assert.Equal(t, "", p.Username)
	assert.Equal(t, "", p.DSL)
}
