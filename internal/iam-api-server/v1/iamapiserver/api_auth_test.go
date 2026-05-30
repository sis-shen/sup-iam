package iamapiserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	servicemock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newTestAuthAPI(ctrl *gomock.Controller) (*AuthAPI, *servicemock.MockAuthCaseInterface, *servicemock.MockUserCaseInterface) {
	mockAuthCase := servicemock.NewMockAuthCaseInterface(ctrl)
	mockUserCase := servicemock.NewMockUserCaseInterface(ctrl)
	jwtMgr := testutil.NewTestJWTManager()
	logger := testutil.NewTestLogger()

	api := NewAuthAPI(logger, jwtMgr, mockAuthCase, mockUserCase)
	return api, mockAuthCase, mockUserCase
}

func TestAuthAPI_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockAuthCase, _ := newTestAuthAPI(ctrl)

	jwtMgr := testutil.NewTestJWTManager()
	mockAuthCase.EXPECT().
		Login(gomock.Any(), gomock.Any()).
		Return(&model.User{ID: 1, Username: "testuser"}, "access-token", "refresh-token", nil)

	body := `{"username":"testuser","password":"Pass1234"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthLoginPost(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
	assert.Equal(t, jwtMgr.GetTokenExpireTime().Microseconds(), resp.ExpiresIn)
}

func TestAuthAPI_Login_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, _ := newTestAuthAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{invalid json}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthLoginPost(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, ParamParseError, errResp.Error)
}

func TestAuthAPI_Login_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockAuthCase, _ := newTestAuthAPI(ctrl)

	mockAuthCase.EXPECT().
		Login(gomock.Any(), gomock.Any()).
		Return(nil, "", "", errors.New("invalid credentials"))

	body := `{"username":"testuser","password":"wrong"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthLoginPost(c)
	// LoginError 返回 200 但 body 中带 error 字段
	assert.Equal(t, http.StatusOK, w.Code)

	var errResp ErrorResponse
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, LoginError, errResp.Error)
}

func TestAuthAPI_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockAuthCase, _ := newTestAuthAPI(ctrl)

	mockAuthCase.EXPECT().
		Register(gomock.Any(), gomock.Any()).
		Return(&model.User{ID: 1, Username: "newuser"}, nil)

	body := `{"username":"newuser","password":"Pass1234"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthRegisterPost(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp RegisterResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.UserId)
	assert.Equal(t, "newuser", resp.Username)
}

func TestAuthAPI_Register_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, _ := newTestAuthAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(`{invalid}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthRegisterPost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthAPI_Register_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockAuthCase, _ := newTestAuthAPI(ctrl)

	mockAuthCase.EXPECT().
		Register(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("username already exists"))

	body := `{"username":"existing","password":"Pass1234"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthRegisterPost(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthAPI_Logout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockAuthCase, _ := newTestAuthAPI(ctrl)

	mockAuthCase.EXPECT().
		Logout(gomock.Any()).
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)

	api.ApiV1AuthLogoutPost(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp LogoutResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "Logout success", resp.Message)
}

func TestAuthAPI_Logout_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockAuthCase, _ := newTestAuthAPI(ctrl)

	mockAuthCase.EXPECT().
		Logout(gomock.Any()).
		Return(errors.New("logout failed"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)

	api.ApiV1AuthLogoutPost(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthAPI_Me_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, mockUserCase := newTestAuthAPI(ctrl)
	jwtMgr := testutil.NewTestJWTManager()

	token, _, err := jwtMgr.GenerateToken("1", "testuser")
	assert.NoError(t, err)

	phone := "13800138000"
	email := "test@example.com"
	mockUserCase.EXPECT().
		GetUserByID(gomock.Any(), "1").
		Return(&model.User{
			ID:       1,
			Username: "testuser",
			Nickname: "Test",
			Phone:    &phone,
			Email:    &email,
			IsAdmin:  1,
			IsEnable: 1,
		}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	// 模拟 middleware 已设置 user_id
	c.Set(UserIDKey, "1")

	api.ApiV1AuthMeGet(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp MeResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, "Test", resp.Nickname)
}

func TestAuthAPI_Me_NoUserIDInContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, _ := newTestAuthAPI(ctrl)
	jwtMgr := testutil.NewTestJWTManager()

	token, _, _ := jwtMgr.GenerateToken("1", "testuser")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)
	// 不设置 user_id — 模拟没有设置的情况

	api.ApiV1AuthMeGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthAPI_ChangePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, mockUserCase := newTestAuthAPI(ctrl)

	mockUserCase.EXPECT().
		GetUserByID(gomock.Any(), "1").
		Return(&model.User{ID: 1, Username: "testuser", PasswordHash: "old-hash"}, nil)
	mockUserCase.EXPECT().
		VerifyPassword("old-pass", "old-hash").
		Return(nil)
	mockUserCase.EXPECT().
		HashPassword("new-pass").
		Return("new-hash", nil)
	mockUserCase.EXPECT().
		UpdateUser(gomock.Any(), gomock.Any()).
		Return(&model.User{ID: 1, PasswordHash: "new-hash"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/change",
		strings.NewReader(`{"old_password":"old-pass","new_password":"new-pass"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthPasswordChangePost(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthAPI_ChangePassword_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, _ := newTestAuthAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/password/change",
		strings.NewReader(`{invalid}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1AuthPasswordChangePost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthAPI_Refresh_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockAuthCase, _ := newTestAuthAPI(ctrl)
	jwtMgr := testutil.NewTestJWTManager()

	// 生成一个 refresh token
	token, _, err := jwtMgr.GenerateToken("1", "testuser")
	assert.NoError(t, err)

	mockAuthCase.EXPECT().
		RefreshToken(gomock.Any(), "1", "testuser").
		Return("new-access", "new-refresh", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	api.ApiV1AuthRefreshPost(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp LoginResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "new-access", resp.AccessToken)
	assert.Equal(t, "new-refresh", resp.RefreshToken)
}

func TestAuthAPI_Refresh_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, _ := newTestAuthAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	c.Request.Header.Set("Authorization", "Bearer invalid-token")

	api.ApiV1AuthRefreshPost(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthAPI_Refresh_NoToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _, _ := newTestAuthAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)

	api.ApiV1AuthRefreshPost(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
