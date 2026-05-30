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

func TestUserCase_GetUserList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	query := repository.PageQuery{Limit: 10}
	expected := repository.PageResult[*model.User]{
		Items: []*model.User{
			{ID: 1, Username: "alice"},
		},
		Total: 1,
	}

	mockRepo.EXPECT().
		GetList(gomock.Any(), query).
		Return(expected, nil)

	result, err := userCase.GetUserList(context.Background(), query)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Items))
	assert.Equal(t, "alice", result.Items[0].Username)
}

func TestUserCase_GetUserList_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	mockRepo.EXPECT().
		GetList(gomock.Any(), gomock.Any()).
		Return(repository.PageResult[*model.User]{}, repository.ErrStorageFailure)

	_, err := userCase.GetUserList(context.Background(), repository.PageQuery{})
	assert.ErrorIs(t, err, repository.ErrStorageFailure)
}

func TestUserCase_GetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	expected := &model.User{ID: 1, Username: "alice"}
	mockRepo.EXPECT().
		GetByID(gomock.Any(), "1").
		Return(expected, nil)

	result, err := userCase.GetUserByID(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestUserCase_GetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	mockRepo.EXPECT().
		GetByID(gomock.Any(), "999").
		Return(nil, repository.ErrNotFound)

	_, err := userCase.GetUserByID(context.Background(), "999")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestUserCase_CreateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	input := &model.User{Username: "newuser"}
	expected := &model.User{ID: 1, Username: "newuser"}

	mockRepo.EXPECT().
		Create(gomock.Any(), input).
		Return(expected, nil)

	result, err := userCase.CreateUser(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestUserCase_UpdateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	input := &model.User{ID: 1, Nickname: "new-nick"}
	expected := &model.User{ID: 1, Nickname: "new-nick"}

	mockRepo.EXPECT().
		Update(gomock.Any(), input).
		Return(expected, nil)

	result, err := userCase.UpdateUser(context.Background(), input)
	assert.NoError(t, err)
	assert.Equal(t, "new-nick", result.Nickname)
}

func TestUserCase_DeleteUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	mockRepo.EXPECT().
		DeleteByID(gomock.Any(), "1").
		Return(nil)

	err := userCase.DeleteUser(context.Background(), "1")
	assert.NoError(t, err)
}

func TestUserCase_DeleteUser_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	mockRepo.EXPECT().
		DeleteByID(gomock.Any(), "999").
		Return(repository.ErrNotFound)

	err := userCase.DeleteUser(context.Background(), "999")
	assert.ErrorIs(t, err, repository.ErrNotFound)
}

func TestUserCase_GetUserByEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	expected := &model.User{ID: 1, Username: "alice"}
	mockRepo.EXPECT().
		GetByEmail(gomock.Any(), "alice@example.com").
		Return(expected, nil)

	result, err := userCase.GetUserByEmail(context.Background(), "alice@example.com")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestUserCase_GetUserByPhone_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	expected := &model.User{ID: 1, Username: "alice"}
	mockRepo.EXPECT().
		GetByPhone(gomock.Any(), "13800138000").
		Return(expected, nil)

	result, err := userCase.GetUserByPhone(context.Background(), "13800138000")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestUserCase_GetUserByUsername_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	userCase := NewUserCase(mockRepo)

	expected := &model.User{ID: 1, Username: "alice"}
	mockRepo.EXPECT().
		GetByUsername(gomock.Any(), "alice").
		Return(expected, nil)

	result, err := userCase.GetUserByUsername(context.Background(), "alice")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}
