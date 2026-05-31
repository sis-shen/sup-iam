package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/cache"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	repomock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// mockBlackList implements cache.TokenBlackListInterface for testing Logout
type mockBlackList struct {
	addFn func(ctx context.Context, token string) error
}

func (m *mockBlackList) Add(ctx context.Context, token string) error {
	return m.addFn(ctx, token)
}

func (m *mockBlackList) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	return false, nil
}

// newTestAuthCase creates an AuthCase with a real hasher and JWT manager for testing
func newTestAuthCase(mockRepo *repomock.MockUserRepository, blackList cache.TokenBlackListInterface) *AuthCase {
	hasher := NewInnerBcryptPasswordHasher(0)
	return NewAuthCase(mockRepo, hasher, testutil.NewTestJWTManager(), blackList)
}

func TestLogin_Success_ByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)

	c := testutil.NewTestGinContext()

	hashedPassword, _ := authCase.hasher.HashPassword("Pass1234")
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
	authCase := newTestAuthCase(mockRepo, nil)

	c := testutil.NewTestGinContext()

	email := "test@example.com"
	hashedPassword, _ := authCase.hasher.HashPassword("Pass1234")
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
	authCase := newTestAuthCase(mockRepo, nil)

	c := testutil.NewTestGinContext()

	phone := "13800138000"
	hashedPassword, _ := authCase.hasher.HashPassword("Pass1234")
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
	authCase := newTestAuthCase(mockRepo, nil)

	c := testutil.NewTestGinContext()

	hashedPassword, _ := authCase.hasher.HashPassword("Pass1234")
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
	authCase := newTestAuthCase(mockRepo, nil)

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
	authCase := newTestAuthCase(mockRepo, nil)

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
	authCase := newTestAuthCase(mockRepo, nil)

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
	authCase := newTestAuthCase(mockRepo, nil)

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
	authCase := newTestAuthCase(mockRepo, nil)

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
	authCase := newTestAuthCase(mockRepo, nil)

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
	authCase := newTestAuthCase(mockRepo, nil)

	c := testutil.NewTestGinContext()

	// Username 非空时，即使 Email/Phone 也有值，应走 GetByUsername
	hashedPassword, _ := authCase.hasher.HashPassword("Pass1234")
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
	authCase := newTestAuthCase(mockRepo, nil)

	c := testutil.NewTestGinContext()

	// Username 为空，Email 非空时，即使 Phone 也有值，应走 GetByEmail
	hashedPassword, _ := authCase.hasher.HashPassword("Pass1234")
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

// ---------------------------------------------------------------------------
// Register 测试
// ---------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	email := "test@example.com"
	phone := "13800138000"
	inputUser := &model.User{
		Username:     "newuser",
		Email:        &email,
		Phone:        &phone,
		PasswordHash: "Pass1234",
	}

	gomock.InOrder(
		mockRepo.EXPECT().ExistsByUsername(c, "newuser").Return(false, nil),
		mockRepo.EXPECT().ExistsByEmail(c, email).Return(false, nil),
		mockRepo.EXPECT().ExistsByPhone(c, phone).Return(false, nil),
	)

	// HashPassword 之后，PasswordHash 会被覆盖，所以 Create 的参数不能直接用 inputUser
	mockRepo.EXPECT().Create(c, gomock.Any()).DoAndReturn(func(_ *gin.Context, u *model.User) (*model.User, error) {
		assert.Equal(t, uint8(1), u.IsEnable)
		assert.Equal(t, uint8(0), u.IsAdmin)
		assert.NotEmpty(t, u.PasswordHash)
		assert.NotEqual(t, "Pass1234", u.PasswordHash, "password should be hashed")
		assert.NotEmpty(t, u.InstanceID)
		assert.False(t, u.CreatedAt.IsZero())
		assert.False(t, u.UpdatedAt.IsZero())
		u.ID = 100
		return u, nil
	})

	result, err := authCase.Register(c, inputUser)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint64(100), result.ID)
}

func TestRegister_DuplicateUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	mockRepo.EXPECT().
		ExistsByUsername(c, "existing").
		Return(true, nil)

	_, err := authCase.Register(c, &model.User{Username: "existing", PasswordHash: "Pass1234"})
	assert.Error(t, err)
	assert.Equal(t, "username already exists", err.Error())
}

func TestRegister_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	email := "dup@example.com"

	gomock.InOrder(
		mockRepo.EXPECT().ExistsByUsername(c, "user").Return(false, nil),
		mockRepo.EXPECT().ExistsByEmail(c, email).Return(true, nil),
	)

	_, err := authCase.Register(c, &model.User{
		Username:     "user",
		Email:        &email,
		PasswordHash: "Pass1234",
	})
	assert.Error(t, err)
	assert.Equal(t, "email already registered", err.Error())
}

func TestRegister_DuplicatePhone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	phone := "13800138000"

	gomock.InOrder(
		mockRepo.EXPECT().ExistsByUsername(c, "user").Return(false, nil),
		mockRepo.EXPECT().ExistsByPhone(c, phone).Return(true, nil),
	)

	_, err := authCase.Register(c, &model.User{
		Username:     "user",
		Phone:        &phone,
		PasswordHash: "Pass1234",
	})
	assert.Error(t, err)
	assert.Equal(t, "phone already registered", err.Error())
}

func TestRegister_WeakPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	mockRepo.EXPECT().
		ExistsByUsername(c, "user").
		Return(false, nil)

	// 无数字
	_, err := authCase.Register(c, &model.User{Username: "user", PasswordHash: "Password"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must contain at least one number")
}

func TestRegister_ExistsByUsername_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	mockRepo.EXPECT().
		ExistsByUsername(c, "user").
		Return(false, errors.New("db error"))

	_, err := authCase.Register(c, &model.User{Username: "user", PasswordHash: "Pass1234"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check username exists failed")
}

func TestRegister_ExistsByEmail_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	email := "err@example.com"

	gomock.InOrder(
		mockRepo.EXPECT().ExistsByUsername(c, "user").Return(false, nil),
		mockRepo.EXPECT().ExistsByEmail(c, email).Return(false, errors.New("db error")),
	)

	_, err := authCase.Register(c, &model.User{
		Username:     "user",
		Email:        &email,
		PasswordHash: "Pass1234",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check email exists failed")
}

func TestRegister_CreateUser_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	gomock.InOrder(
		mockRepo.EXPECT().ExistsByUsername(c, "newuser").Return(false, nil),
	)

	mockRepo.EXPECT().Create(c, gomock.Any()).Return(nil, errors.New("db insert failed"))

	_, err := authCase.Register(c, &model.User{Username: "newuser", PasswordHash: "Pass1234"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create user failed")
}

func TestRegister_RegisterWithoutEmailAndPhone(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)
	c := testutil.NewTestGinContext()

	// 只有用户名，没有邮箱和手机，应只检查用户名
	mockRepo.EXPECT().ExistsByUsername(c, "minimal").Return(false, nil)
	mockRepo.EXPECT().Create(c, gomock.Any()).DoAndReturn(func(_ *gin.Context, u *model.User) (*model.User, error) {
		u.ID = 1
		return u, nil
	})

	result, err := authCase.Register(c, &model.User{Username: "minimal", PasswordHash: "Pass1234"})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
}

// ---------------------------------------------------------------------------
// Logout 测试
// ---------------------------------------------------------------------------

func TestLogout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	blackList := &mockBlackList{
		addFn: func(_ context.Context, token string) error {
			assert.Equal(t, "my-token", token)
			return nil
		},
	}
	authCase := newTestAuthCase(mockRepo, blackList)

	// 使用 gin context 并设置 Authorization header，以便 ExtractToken 能提取到 token
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	c.Request.Header.Set("Authorization", "Bearer my-token")

	err := authCase.Logout(c)
	assert.NoError(t, err)
}

func TestLogout_NoToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)

	// 无 Authorization header
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)

	err := authCase.Logout(c)
	assert.Error(t, err)
	assert.Equal(t, "fail to extract token", err.Error())
}

func TestLogout_BlackListAddError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	blackList := &mockBlackList{
		addFn: func(_ context.Context, token string) error {
			return errors.New("redis error")
		},
	}
	authCase := newTestAuthCase(mockRepo, blackList)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	c.Request.Header.Set("Authorization", "Bearer my-token")

	err := authCase.Logout(c)
	assert.Error(t, err)
	assert.Equal(t, "redis error", err.Error())
}

// ---------------------------------------------------------------------------
// RefreshToken 测试
// ---------------------------------------------------------------------------

func TestRefreshToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	c.Request.Header.Set("Authorization", "Bearer existing-token")

	accessToken, refreshToken, err := authCase.RefreshToken(c, "user-001", "alice")
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestRefreshToken_NoToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/refresh", nil)
	// no Authorization header

	_, _, err := authCase.RefreshToken(c, "user-001", "alice")
	assert.Error(t, err)
	assert.Equal(t, "fail to extract token", err.Error())
}

// ---------------------------------------------------------------------------
// 并发测试
// ---------------------------------------------------------------------------

func TestRegister_ConcurrentDuplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockUserRepository(ctrl)
	authCase := newTestAuthCase(mockRepo, nil)

	c := testutil.NewTestGinContext()

	username := "concurrent-user"
	password := "Pass1234"
	n := 10

	// 模拟真实竞态场景：所有请求都通过 ExistsByUsername 检查，
	// 但在 Create 时只有第 1 个成功，其余触发 MySQL 唯一索引冲突（1062）
	var createCount atomic.Int32

	// ExistsByUsername: 全部返回 false
	mockRepo.EXPECT().
		ExistsByUsername(c, username).
		Return(false, nil).
		Times(n)

	// Create: 第 1 次成功，后续返回 1062 错误（模拟数据库唯一索引拦截）
	mockRepo.EXPECT().
		Create(c, gomock.Any()).
		DoAndReturn(func(_ *gin.Context, u *model.User) (*model.User, error) {
			if createCount.Add(1) == 1 {
				u.ID = 200
				return u, nil
			}
			return nil, repository.ErrAlreadyExists
		}).
		Times(n)

	var wg sync.WaitGroup
	results := make(chan error, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := authCase.Register(c, &model.User{
				Username:     username,
				PasswordHash: password,
			})
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	successCount := 0
	errorCount := 0
	for err := range results {
		if err == nil {
			successCount++
		} else {
			errorCount++
		}
	}

	// 只有 1 个能创建成功，其余都因 DB 唯一约束触发 "create user failed: ..."
	assert.Equal(t, 1, successCount, "exactly one registration should succeed")
	assert.Equal(t, n-1, errorCount, "the rest %d should fail", n-1)
}
