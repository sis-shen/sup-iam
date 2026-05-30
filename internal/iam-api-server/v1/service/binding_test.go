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

func TestBindingCase_GetBindingListByUserID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockBindingRepository(ctrl)
	bindingCase := NewBindingCase(mockRepo)

	query := repository.PageQuery{Limit: 10}
	expected := repository.PageResult[*model.Binding]{
		Items: []*model.Binding{{ID: 1, SecretID: 1, PolicyID: 1}},
		Total: 1,
	}

	mockRepo.EXPECT().
		GetListByUserID(gomock.Any(), "1", query).
		Return(expected, nil)

	result, err := bindingCase.GetBindingListByUserID(context.Background(), "1", query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
}

func TestBindingCase_GetBindingById_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockBindingRepository(ctrl)
	bindingCase := NewBindingCase(mockRepo)

	expected := &model.Binding{ID: 1, SecretID: 1, PolicyID: 1}
	mockRepo.EXPECT().
		GetByID(gomock.Any(), "1").
		Return(expected, nil)

	result, err := bindingCase.GetBindingById(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestBindingCase_GetBindingById_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockBindingRepository(ctrl)
	bindingCase := NewBindingCase(mockRepo)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), "999").
		Return(nil, repository.ErrNotFound)

	_, err := bindingCase.GetBindingById(context.Background(), "999")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestBindingCase_CreateBinding_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockBindingRepository(ctrl)
	bindingCase := NewBindingCase(mockRepo)

	input := &model.Binding{SecretID: 1, PolicyID: 1}
	expected := &model.Binding{ID: 1, SecretID: 1, PolicyID: 1}

	mockRepo.EXPECT().
		Create(gomock.Any(), input).
		Return(expected, nil)

	result, err := bindingCase.CreateBinding(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestBindingCase_DeleteBinding_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockBindingRepository(ctrl)
	bindingCase := NewBindingCase(mockRepo)

	mockRepo.EXPECT().
		DeleteByID(gomock.Any(), "1").
		Return(nil)

	err := bindingCase.DeleteBinding(context.Background(), "1")
	assert.NoError(t, err)
}

func TestBindingCase_DeleteBinding_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockBindingRepository(ctrl)
	bindingCase := NewBindingCase(mockRepo)

	mockRepo.EXPECT().
		DeleteByID(gomock.Any(), "999").
		Return(repository.ErrNotFound)

	err := bindingCase.DeleteBinding(context.Background(), "999")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}
