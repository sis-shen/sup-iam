package service

import (
	"errors"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	repomock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestLogin_Success_ByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	hashedPassword, _ := hasher.HashPassword("Pass1234")
	dbUser := &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hashedPassword,
	}

	mockRepo.EXPECT().
		GetByUsername(c, "testuser").
		Return(dbUser, nil)

	inputUser := &model.User{
		Username:     "testuser",
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "testuser", result.Username)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken, "access token and refresh token should differ")
}

func TestLogin_Success_ByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	email := "test@example.com"
	hashedPassword, _ := hasher.HashPassword("Pass1234")
	dbUser := &model.User{
		ID:           2,
		Username:     "testuser",
		Email:        &email,
		PasswordHash: hashedPassword,
	}

	mockRepo.EXPECT().
		GetByEmail(c, email).
		Return(dbUser, nil)

	inputUser := &model.User{
		Email:        &email,
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(2), result.ID)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestLogin_Success_ByPhone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	phone := "13800138000"
	hashedPassword, _ := hasher.HashPassword("Pass1234")
	dbUser := &model.User{
		ID:           3,
		Username:     "testuser",
		Phone:        &phone,
		PasswordHash: hashedPassword,
	}

	mockRepo.EXPECT().
		GetByPhone(c, phone).
		Return(dbUser, nil)

	inputUser := &model.User{
		Phone:        &phone,
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(3), result.ID)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestLogin_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	hashedPassword, _ := hasher.HashPassword("Pass1234")
	dbUser := &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hashedPassword,
	}

	mockRepo.EXPECT().
		GetByUsername(c, "testuser").
		Return(dbUser, nil)

	inputUser := &model.User{
		Username:     "testuser",
		PasswordHash: "WrongPassword1",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.Error(t, err)
	assert.Equal(t, "invalid password", err.Error())
	assert.Nil(t, result)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_UserNotFound_ByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	mockRepo.EXPECT().
		GetByUsername(c, "nonexistent").
		Return(nil, repository.ErrNotFound)

	inputUser := &model.User{
		Username:     "nonexistent",
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user by username failed")
	assert.Nil(t, result)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_DbUserIsNil_ReturnsUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	// repo 返回 (nil, nil) 而非 error
	mockRepo.EXPECT().
		GetByUsername(c, "ghost").
		Return(nil, nil)

	inputUser := &model.User{
		Username:     "ghost",
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, result)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_NoIdentifierProvided_ReturnsUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	// all identifiers empty — repo should NOT be called
	inputUser := &model.User{
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
	assert.Nil(t, result)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_GetByUsername_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	mockRepo.EXPECT().
		GetByUsername(c, "testuser").
		Return(nil, errors.New("db connection timeout"))

	inputUser := &model.User{
		Username:     "testuser",
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user by username failed")
	assert.Nil(t, result)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_GetByEmail_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	email := "test@example.com"
	mockRepo.EXPECT().
		GetByEmail(c, email).
		Return(nil, errors.New("db connection timeout"))

	inputUser := &model.User{
		Email:        &email,
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user by email failed")
	assert.Nil(t, result)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_GetByPhone_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	phone := "13800138000"
	mockRepo.EXPECT().
		GetByPhone(c, phone).
		Return(nil, errors.New("db connection timeout"))

	inputUser := &model.User{
		Phone:        &phone,
		PasswordHash: "Pass1234",
	}

	result, accessToken, refreshToken, err := authCase.Login(c, inputUser)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get user by phone failed")
	assert.Nil(t, result)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
}

func TestLogin_UsernameTakesPriorityOverEmailAndPhone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	// Username 非空时，即使 Email/Phone 也有值，应走 GetByUsername
	hashedPassword, _ := hasher.HashPassword("Pass1234")
	dbUser := &model.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: hashedPassword,
	}

	mockRepo.EXPECT().
		GetByUsername(c, "testuser").
		Return(dbUser, nil)

	email := "should@be.ignored"
	inputUser := &model.User{
		Username:     "testuser",
		Email:        &email,
		PasswordHash: "Pass1234",
	}

	result, _, _, err := authCase.Login(c, inputUser)

	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

func TestLogin_EmailTakesPriorityOverPhone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	hasher := NewInnerBcryptPasswordHasher(0)
	authCase := NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), nil)

	c := testutil.NewTestGinContext()

	// Username 为空，Email 非空时，即使 Phone 也有值，应走 GetByEmail
	hashedPassword, _ := hasher.HashPassword("Pass1234")
	email := "test@example.com"
	dbUser := &model.User{
		ID:           2,
		Username:     "testuser",
		Email:        &email,
		PasswordHash: hashedPassword,
	}

	mockRepo.EXPECT().
		GetByEmail(c, email).
		Return(dbUser, nil)

	phone := "13800138000"
	inputUser := &model.User{
		Email:        &email,
		Phone:        &phone,
		PasswordHash: "Pass1234",
	}

	result, _, _, err := authCase.Login(c, inputUser)

	assert.NoError(t, err)
	assert.Equal(t, uint64(2), result.ID)
}
