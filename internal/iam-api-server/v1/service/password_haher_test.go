package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestHashAndVerify_Success(t *testing.T) {
	hasher := NewInnerBcryptPasswordHasher(4) // 低cost加速测试

	hash, err := hasher.HashPassword("Pass1234")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = hasher.VerifyPassword("Pass1234", hash)
	assert.NoError(t, err)
}

func TestVerify_WrongPassword(t *testing.T) {
	hasher := NewInnerBcryptPasswordHasher(4)

	hash, err := hasher.HashPassword("Pass1234")
	assert.NoError(t, err)

	err = hasher.VerifyPassword("WrongPassword1", hash)
	assert.Error(t, err)
}

func TestVerify_InvalidHash(t *testing.T) {
	hasher := NewInnerBcryptPasswordHasher(4)

	err := hasher.VerifyPassword("Pass1234", "not-a-valid-hash")
	assert.Error(t, err)
}

func TestNewInnerBcryptPasswordHasher_DefaultCost(t *testing.T) {
	hasher := NewInnerBcryptPasswordHasher(0)
	assert.NotNil(t, hasher)
	assert.Equal(t, bcrypt.DefaultCost, hasher.cost)
}

func TestNewInnerBcryptPasswordHasher_CustomCost(t *testing.T) {
	hasher := NewInnerBcryptPasswordHasher(8)
	assert.NotNil(t, hasher)
	assert.Equal(t, 8, hasher.cost)
}

func TestHash_DifferentSaltEachTime(t *testing.T) {
	hasher := NewInnerBcryptPasswordHasher(4)

	hash1, _ := hasher.HashPassword("Pass1234")
	hash2, _ := hasher.HashPassword("Pass1234")

	assert.NotEqual(t, hash1, hash2, "bcrypt should use different salt each time")
}
