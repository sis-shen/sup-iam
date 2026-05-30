package service

import (
	"context"
	"testing"

	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	repomock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPolicyCase_GetPolicyList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	query := repository.PageQuery{Limit: 10}
	expected := repository.PageResult[*model.Policy]{
		Items: []*model.Policy{{ID: 1, Name: "policy-1"}},
		Total: 1,
	}

	mockRepo.EXPECT().
		GetListByUserID(gomock.Any(), "1", query).
		Return(expected, nil)

	result, err := policyCase.GetPolicyList(context.Background(), "1", query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
}

func TestPolicyCase_GetPolicyByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	expected := &model.Policy{ID: 1, Name: "policy-1"}
	mockRepo.EXPECT().
		GetByID(gomock.Any(), "1").
		Return(expected, nil)

	result, err := policyCase.GetPolicyByID(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, "policy-1", result.Name)
}

func TestPolicyCase_GetPolicyByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), "999").
		Return(nil, repository.ErrNotFound)

	_, err := policyCase.GetPolicyByID(context.Background(), "999")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPolicyCase_CreatePolicy_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	input := &model.Policy{Name: "new-policy"}
	expected := &model.Policy{ID: 1, Name: "new-policy"}

	mockRepo.EXPECT().
		Create(gomock.Any(), input).
		Return(expected, nil)

	result, err := policyCase.CreatePolicy(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestPolicyCase_UpdatePolicy_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	input := &model.Policy{ID: 1, Name: "updated-name"}

	mockRepo.EXPECT().
		Update(gomock.Any(), input).
		Return(input, nil)

	result, err := policyCase.UpdatePolicy(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "updated-name", result.Name)
}

func TestPolicyCase_DeletePolicy_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	mockRepo.EXPECT().
		DeleteByID(gomock.Any(), "1").
		Return(nil)

	err := policyCase.DeletePolicy(context.Background(), "1")
	assert.NoError(t, err)
}

func TestPolicyCase_DeletePolicy_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	mockRepo.EXPECT().
		DeleteByID(gomock.Any(), "999").
		Return(repository.ErrNotFound)

	err := policyCase.DeletePolicy(context.Background(), "999")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPolicyCase_GetPolicyBindingSecretList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockPolicyRepository(ctrl)
	policyCase := NewPolicyCase(mockRepo)

	query := repository.PageQuery{Limit: 10}
	expected := repository.PageResult[*model.Secret]{
		Items: []*model.Secret{{ID: 1, AccessKey: "ak1"}},
		Total: 1,
	}

	mockRepo.EXPECT().
		GetSecretListByPolicyID(gomock.Any(), "1", query).
		Return(expected, nil)

	result, err := policyCase.GetPolicyBindingSecretList(context.Background(), "1", query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
	assert.Equal(t, "ak1", result.Items[0].AccessKey)
}
