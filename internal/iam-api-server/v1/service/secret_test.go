package service

import (
	"context"
	"testing"

	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	repomock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/keys"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSecretCase_RotateSecret_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	k := keys.NewKeys(0, 0)
	secretCase := NewSecretCase(mockRepo)
	secretCase.keys = k

	oldSecret := &model.Secret{
		ID:        1,
		SecretKey: "old-secret-key",
		AccessKey: "access-key-1",
	}

	mockRepo.EXPECT().
		GetByID(gomock.Any(), "1").
		Return(oldSecret, nil)

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Do(func(_ context.Context, s *model.Secret) {
			assert.Equal(t, uint64(1), s.ID)
			assert.NotEqual(t, "old-secret-key", s.SecretKey, "secret key should be rotated")
			assert.NotEmpty(t, s.SecretKey)
		}).
		Return(&model.Secret{ID: 1, SecretKey: "new-secret-key"}, nil)

	result, err := secretCase.RotateSecret(context.Background(), "1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new-secret-key", result.SecretKey)
}

func TestSecretCase_RotateSecret_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	secretCase := NewSecretCase(mockRepo)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), "999").
		Return(nil, repository.ErrNotFound)

	_, err := secretCase.RotateSecret(context.Background(), "999")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestSecretCase_GetSecretList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	secretCase := NewSecretCase(mockRepo)

	query := repository.PageQuery{Limit: 10}
	expected := repository.PageResult[*model.Secret]{
		Items: []*model.Secret{
			{ID: 1, AccessKey: "ak1"},
		},
		Total: 1,
	}

	mockRepo.EXPECT().
		GetListByUserID(gomock.Any(), "1", query).
		Return(expected, nil)

	result, err := secretCase.GetSecretList(context.Background(), "1", query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
}

func TestSecretCase_GetSecretByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	secretCase := NewSecretCase(mockRepo)

	expected := &model.Secret{ID: 1, AccessKey: "ak1"}
	mockRepo.EXPECT().
		GetByID(gomock.Any(), "1").
		Return(expected, nil)

	result, err := secretCase.GetSecretByID(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestSecretCase_GetSecretByAK_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	secretCase := NewSecretCase(mockRepo)

	expected := &model.Secret{ID: 1, AccessKey: "ak1", SecretKey: "sk1"}
	mockRepo.EXPECT().
		GetByAK(gomock.Any(), "ak1").
		Return(expected, nil)

	result, err := secretCase.GetSecretByAK(context.Background(), "ak1")
	assert.NoError(t, err)
	assert.Equal(t, "sk1", result.SecretKey)
}

func TestSecretCase_CreateSecret_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	secretCase := NewSecretCase(mockRepo)

	input := &model.Secret{AccessKey: "new-ak"}
	expected := &model.Secret{ID: 1, AccessKey: "new-ak"}

	mockRepo.EXPECT().
		Create(gomock.Any(), input).
		Return(expected, nil)

	result, err := secretCase.CreateSecret(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestSecretCase_DeleteSecret_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	secretCase := NewSecretCase(mockRepo)

	mockRepo.EXPECT().
		DeleteByID(gomock.Any(), "1").
		Return(nil)

	err := secretCase.DeleteSecret(context.Background(), "1")
	assert.NoError(t, err)
}

func TestSecretCase_GetSecretBindingPolicy_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockSecretRepository(ctrl)
	secretCase := NewSecretCase(mockRepo)

	query := repository.PageQuery{Limit: 10}
	expected := repository.PageResult[*model.Policy]{
		Items: []*model.Policy{
			{ID: 1, Name: "policy-1"},
		},
		Total: 1,
	}

	mockRepo.EXPECT().
		GetPolicyListBySecretID(gomock.Any(), "1", query).
		Return(expected, nil)

	result, err := secretCase.GetSecretBindingPolicy(context.Background(), "1", query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
	assert.Equal(t, "policy-1", result.Items[0].Name)
}

func TestSecretCase_VerifySecret_Valid(t *testing.T) {
	k := keys.NewKeys(0, 0)
	secretCase := NewSecretCase(nil)
	secretCase.keys = k

	payload := "test-payload"
	sig, _ := k.SignWithKey("my-secret", payload)

	ok, err := secretCase.VerifySecret("my-secret", payload, sig)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestSecretCase_VerifySecret_Invalid(t *testing.T) {
	k := keys.NewKeys(0, 0)
	secretCase := NewSecretCase(nil)
	secretCase.keys = k

	ok, err := secretCase.VerifySecret("my-secret", "payload", "wrong-signature")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestSecretCase_GenerateSecretKey_NotEmpty(t *testing.T) {
	k := keys.NewKeys(0, 0)
	secretCase := NewSecretCase(nil)
	secretCase.keys = k

	sk := secretCase.GenerateSecretKey()
	assert.NotEmpty(t, sk)
}

func TestSecretCase_GenerateAccessKey_NotEmpty(t *testing.T) {
	k := keys.NewKeys(0, 0)
	secretCase := NewSecretCase(nil)
	secretCase.keys = k

	ak := secretCase.GenerateAccessKey()
	assert.NotEmpty(t, ak)
}
