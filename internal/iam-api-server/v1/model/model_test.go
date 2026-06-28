package model

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserStruct(t *testing.T) {
	now := time.Now()
	phone := "13800138000"
	email := "test@example.com"
	ext := "{\"key\":\"value\"}"
	loggedAt := now

	u := &User{
		ID:           1,
		InstanceID:   "instance-001",
		Username:     "testuser",
		Nickname:     "Test User",
		PasswordHash: "hashed-password",
		IsEnable:     1,
		Phone:        &phone,
		Email:        &email,
		IsAdmin:      0,
		ExtendShadow: &ext,
		LoggedAt:     &loggedAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, uint64(1), u.ID)
	assert.Equal(t, "testuser", u.Username)
	assert.Equal(t, "Test User", u.Nickname)
	assert.Equal(t, "hashed-password", u.PasswordHash)
	assert.Equal(t, uint8(1), u.IsEnable)
	assert.Equal(t, uint8(0), u.IsAdmin)
	assert.NotNil(t, u.Phone)
	assert.Equal(t, "13800138000", *u.Phone)
	assert.NotNil(t, u.Email)
	assert.Equal(t, "test@example.com", *u.Email)
	assert.NotNil(t, u.LoggedAt)
	assert.Equal(t, now, *u.LoggedAt)
}

func TestUser_ZeroValues(t *testing.T) {
	u := &User{}
	assert.Equal(t, uint64(0), u.ID)
	assert.Equal(t, "", u.Username)
	assert.Equal(t, uint8(0), u.IsEnable)
	assert.Nil(t, u.Phone)
	assert.Nil(t, u.Email)
	assert.Nil(t, u.LoggedAt)
	assert.Nil(t, u.ExtendShadow)
}

func TestPolicyStruct(t *testing.T) {
	now := time.Now()
	desc := "Test policy description"
	dsl := "p, user, resource, action"
	ext := "{}"

	p := &Policy{
		ID:           1,
		InstanceID:   "policy-001",
		Name:         "test-policy",
		Username:     "testuser",
		Description:  &desc,
		PolicyShadow: &dsl,
		ExtendShadow: &ext,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, uint64(1), p.ID)
	assert.Equal(t, "test-policy", p.Name)
	assert.Equal(t, "testuser", p.Username)
	assert.NotNil(t, p.Description)
	assert.Equal(t, "Test policy description", *p.Description)
	assert.NotNil(t, p.PolicyShadow)
	assert.Equal(t, "p, user, resource, action", *p.PolicyShadow)
}

func TestPolicy_ZeroValues(t *testing.T) {
	p := &Policy{}
	assert.Equal(t, uint64(0), p.ID)
	assert.Equal(t, "", p.Name)
	assert.Nil(t, p.Description)
	assert.Nil(t, p.PolicyShadow)
	assert.Nil(t, p.ExtendShadow)
}

func TestSecretStruct(t *testing.T) {
	now := time.Now()
	desc := "My API key"
	ext := "{}"

	s := &Secret{
		ID:           1,
		InstanceID:   "secret-001",
		UserID:       100,
		Username:     "testuser",
		AccessKey:    "AKID123456",
		SecretKey:    "SK123456",
		Expires:      1735689600,
		Description:  &desc,
		ExtendShadow: &ext,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, uint64(1), s.ID)
	assert.Equal(t, uint64(100), s.UserID)
	assert.Equal(t, "AKID123456", s.AccessKey)
	assert.Equal(t, "SK123456", s.SecretKey)
	assert.Equal(t, int64(1735689600), s.Expires)
	assert.NotNil(t, s.Description)
	assert.Equal(t, "My API key", *s.Description)
}

func TestBindingStruct(t *testing.T) {
	now := time.Now()
	ext := "{}"

	b := &Binding{
		ID:           1,
		SecretID:     100,
		PolicyID:     200,
		Username:     "testuser",
		ExtendShadow: &ext,
		CreatedAt:    now,
	}

	assert.Equal(t, uint64(1), b.ID)
	assert.Equal(t, uint64(100), b.SecretID)
	assert.Equal(t, uint64(200), b.PolicyID)
	assert.Equal(t, "testuser", b.Username)
	assert.NotNil(t, b.ExtendShadow)
}

func TestBinding_ZeroValues(t *testing.T) {
	b := &Binding{}
	assert.Equal(t, uint64(0), b.ID)
	assert.Equal(t, uint64(0), b.SecretID)
	assert.Equal(t, uint64(0), b.PolicyID)
	assert.Equal(t, "", b.Username)
	assert.Nil(t, b.ExtendShadow)
}

func TestPolicyAuditStruct(t *testing.T) {
	now := time.Now()
	desc := "Audit entry"
	dsl := "p, user, res, act"

	pa := &PolicyAudit{
		ID:           1,
		InstanceID:   "audit-001",
		Name:         "policy-1",
		Username:     "admin",
		Description:  &desc,
		PolicyShadow: &dsl,
		ExtendShadow: nil,
		CreatedAt:    now,
	}

	assert.Equal(t, uint64(1), pa.ID)
	assert.Equal(t, "policy-1", pa.Name)
	assert.Equal(t, "admin", pa.Username)
	assert.NotNil(t, pa.Description)
	assert.Equal(t, "Audit entry", *pa.Description)
	assert.NotNil(t, pa.PolicyShadow)
	assert.Equal(t, "p, user, res, act", *pa.PolicyShadow)
	assert.Nil(t, pa.ExtendShadow)
}

func TestBindingAuditStruct(t *testing.T) {
	now := time.Now()
	ext := "{}"

	ba := &BindingAudit{
		ID:           1,
		SecretID:     100,
		PolicyID:     200,
		Username:     "admin",
		ExtendShadow: &ext,
		CreatedAt:    now,
	}

	assert.Equal(t, uint64(1), ba.ID)
	assert.Equal(t, uint64(100), ba.SecretID)
	assert.Equal(t, uint64(200), ba.PolicyID)
	assert.Equal(t, "admin", ba.Username)
	assert.NotNil(t, ba.ExtendShadow)
}

func TestBindingAudit_ZeroValues(t *testing.T) {
	ba := &BindingAudit{}
	assert.Equal(t, uint64(0), ba.ID)
	assert.Equal(t, uint64(0), ba.SecretID)
	assert.Equal(t, uint64(0), ba.PolicyID)
	assert.Equal(t, "", ba.Username)
	assert.Nil(t, ba.ExtendShadow)
}

func TestPolicyAudit_DoesNotHaveUpdatedAt(t *testing.T) {
	// Use reflection to verify PolicyAudit does not have UpdatedAt field.
	// Zero-value checks cannot distinguish "field doesn't exist" from "field has zero value",
	// so we inspect the struct type directly.
	typ := reflect.TypeOf(PolicyAudit{})
	for i := range typ.NumField() {
		if typ.Field(i).Name == "UpdatedAt" {
			t.Error("PolicyAudit should not have UpdatedAt field (it's an audit/history record)")
		}
	}
}
