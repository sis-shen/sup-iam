package iamapiserver

import (
	"testing"
	"time"

	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestParseUserModel(t *testing.T) {
	now := time.Now()
	phone := "13800138000"
	email := "test@example.com"
	extShadow := "{}"
	user := model.User{
		ID:           1,
		InstanceID:   "inst-001",
		Username:     "testuser",
		Nickname:     "Test",
		IsEnable:     1,
		Phone:        &phone,
		Email:        &email,
		IsAdmin:      1,
		ExtendShadow: &extShadow,
		LoggedAt:     &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := parseUserModel(user)
	assert.Equal(t, int64(1), result.Id)
	assert.Equal(t, "inst-001", result.InstanceId)
	assert.Equal(t, "testuser", result.Username)
	assert.Equal(t, "Test", result.Nickname)
	assert.Equal(t, int32(1), result.IsEnable)
	assert.Equal(t, &phone, result.Phone)
	assert.Equal(t, &email, result.Email)
	assert.Equal(t, int32(1), result.IsAdmin)
	assert.Equal(t, &extShadow, result.ExtendShadow)
	assert.Equal(t, &now, result.LoggedAt)
}

func TestGetUserModel(t *testing.T) {
	now := time.Now()
	phone := "13800138000"
	email := "test@example.com"
	extShadow := "{}"
	user := User{
		Id:           1,
		InstanceId:   "inst-001",
		Username:     "testuser",
		Nickname:     "Test",
		IsEnable:     1,
		Phone:        &phone,
		Email:        &email,
		IsAdmin:      1,
		ExtendShadow: &extShadow,
		LoggedAt:     &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := getUserModel(user)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "inst-001", result.InstanceID)
	assert.Equal(t, uint8(1), result.IsEnable)
	assert.Equal(t, uint8(1), result.IsAdmin)
}

func TestParseUserModelList_Success(t *testing.T) {
	users := []*model.User{
		{ID: 1, Username: "alice"},
		{ID: 2, Username: "bob"},
	}

	result, err := parseUserModelList(users)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
	assert.Equal(t, "alice", result[0].Username)
	assert.Equal(t, "bob", result[1].Username)
}

func TestParseUserModelList_Empty(t *testing.T) {
	result, err := parseUserModelList([]*model.User{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestParseSecretModel(t *testing.T) {
	now := time.Now()
	desc := "my secret"
	extShadow := "{}"
	expires := int64(3600)
	secret := model.Secret{
		ID:           1,
		InstanceID:   "inst-001",
		UserID:       10,
		Username:     "testuser",
		AccessKey:    "ak-001",
		Description:  &desc,
		Expires:      expires,
		ExtendShadow: &extShadow,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := ParseSecretModel(secret)
	assert.Equal(t, int64(1), result.Id)
	assert.Equal(t, "ak-001", result.AccessKey)
	assert.Equal(t, &desc, result.Description)
}

func TestParseSecretModelList_Empty(t *testing.T) {
	result, err := ParseSecretModelList([]*model.Secret{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestParsePolicyModel(t *testing.T) {
	now := time.Now()
	desc := "my policy"
	content := "{}"
	policy := model.Policy{
		ID:           1,
		InstanceID:   "inst-001",
		Name:         "policy-1",
		Description:  &desc,
		PolicyShadow: &content,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	result := ParsePolicyModel(policy)
	assert.Equal(t, int64(1), result.Id)
	assert.Equal(t, "policy-1", result.Name)
	assert.Equal(t, content, result.Content)
}

func TestParsePolicyModelList_Empty(t *testing.T) {
	result, err := ParsePolicyModelList([]*model.Policy{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestParseBindingModel(t *testing.T) {
	now := time.Now()
	binding := model.Binding{
		ID:        1,
		SecretID:  10,
		PolicyID:  20,
		Username:  "testuser",
		CreatedAt: now,
	}

	result := ParseBindingModel(binding)
	assert.Equal(t, int64(1), result.BindingId)
	assert.Equal(t, int64(10), result.SecretId)
	assert.Equal(t, int64(20), result.PolicyId)
	assert.Equal(t, "testuser", result.Username)
}

func TestParseBindingModelList_Empty(t *testing.T) {
	result, err := ParseBindingModelList([]*model.Binding{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestUserIDKey_Value(t *testing.T) {
	assert.Equal(t, "user_id", UserIDKey)
}
