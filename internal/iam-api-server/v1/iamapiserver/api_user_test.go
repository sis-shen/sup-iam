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
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	servicemock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newTestUserAPI(ctrl *gomock.Controller) (*UserAPI, *servicemock.MockUserCaseInterface) {
	mockUserCase := servicemock.NewMockUserCaseInterface(ctrl)
	logger := testutil.NewTestLogger()
	api := NewUserAPI(mockUserCase, logger)
	return api, mockUserCase
}

func TestUserAPI_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockUserCase := newTestUserAPI(ctrl)

	mockUserCase.EXPECT().
		GetUserList(gomock.Any(), gomock.Any()).
		Return(repository.PageResult[*model.User]{
			Items: []*model.User{
				{ID: 1, Username: "alice"},
				{ID: 2, Username: "bob"},
			},
			Total: 2,
		}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&page_size=20", nil)

	api.ApiV1UsersGet(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp UserListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, len(resp.Items))
	assert.Equal(t, int64(2), resp.Total)
}

func TestUserAPI_List_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockUserCase := newTestUserAPI(ctrl)

	mockUserCase.EXPECT().
		GetUserList(gomock.Any(), gomock.Any()).
		Return(repository.PageResult[*model.User]{}, errors.New("database error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)

	api.ApiV1UsersGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserAPI_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockUserCase := newTestUserAPI(ctrl)

	mockUserCase.EXPECT().
		GetUserByID(gomock.Any(), "1").
		Return(&model.User{ID: 1, Username: "alice"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1UsersIdGet(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp User
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, "alice", resp.Username)
}

func TestUserAPI_GetByID_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestUserAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/", nil)
	// 不设置 id param

	api.ApiV1UsersIdGet(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAPI_GetByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockUserCase := newTestUserAPI(ctrl)

	mockUserCase.EXPECT().
		GetUserByID(gomock.Any(), "999").
		Return(nil, errors.New("not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/users/999", nil)
	c.Params = []gin.Param{{Key: "id", Value: "999"}}

	api.ApiV1UsersIdGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUserAPI_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockUserCase := newTestUserAPI(ctrl)

	mockUserCase.EXPECT().
		DeleteUser(gomock.Any(), "1").
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1UsersIdDelete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserAPI_Delete_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestUserAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/users/", nil)

	api.ApiV1UsersIdDelete(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAPI_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockUserCase := newTestUserAPI(ctrl)

	mockUserCase.EXPECT().
		HashPassword("Pass1234").
		Return("hashed-pass", nil)
	mockUserCase.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(&model.User{ID: 1, Username: "newuser"}, nil)

	body := `{"username":"newuser","password":"Pass1234","nickname":"New"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1UsersPost(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp User
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, "newuser", resp.Username)
}

func TestUserAPI_Create_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestUserAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(`{invalid}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1UsersPost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAPI_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockUserCase := newTestUserAPI(ctrl)

	mockUserCase.EXPECT().
		GetUserByID(gomock.Any(), "1").
		Return(&model.User{ID: 1, Username: "alice", Nickname: "old-nick"}, nil)
	mockUserCase.EXPECT().
		UpdateUser(gomock.Any(), gomock.Any()).
		Return(&model.User{ID: 1, Username: "alice", Nickname: "new-nick"}, nil)

	body := `{"nickname":"new-nick"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1UsersIdPut(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp User
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
}

func TestUserAPI_Update_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestUserAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1UsersIdPut(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserAPI_Update_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestUserAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/1", strings.NewReader(`{invalid}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1UsersIdPut(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
