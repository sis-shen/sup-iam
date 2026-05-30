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

func newTestBindingAPI(ctrl *gomock.Controller) (*BindingAPI, *servicemock.MockBindingCaseInterface) {
	mockBindingCase := servicemock.NewMockBindingCaseInterface(ctrl)
	logger := testutil.NewTestLogger()
	api := NewBindingAPI(mockBindingCase, logger)
	return api, mockBindingCase
}

func TestBindingAPI_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockBindingCase := newTestBindingAPI(ctrl)

	mockBindingCase.EXPECT().
		GetBindingListByUserID(gomock.Any(), "1", gomock.Any()).
		Return(repository.PageResult[*model.Binding]{
			Items: []*model.Binding{
				{ID: 1, SecretID: 10, PolicyID: 20},
				{ID: 2, SecretID: 11, PolicyID: 21},
			},
			Total: 2,
		}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/bindings?page=1&page_size=20", nil)

	api.ApiV1BindingsGet(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BindingListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, len(resp.Items))
	assert.Equal(t, int64(2), resp.Total)
}

func TestBindingAPI_List_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestBindingAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 不设置 UserIDKey
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/bindings", nil)

	api.ApiV1BindingsGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBindingAPI_List_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockBindingCase := newTestBindingAPI(ctrl)

	mockBindingCase.EXPECT().
		GetBindingListByUserID(gomock.Any(), "1", gomock.Any()).
		Return(repository.PageResult[*model.Binding]{}, errors.New("database error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/bindings", nil)

	api.ApiV1BindingsGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestBindingAPI_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockBindingCase := newTestBindingAPI(ctrl)

	mockBindingCase.EXPECT().
		GetBindingById(gomock.Any(), "1").
		Return(&model.Binding{ID: 1, SecretID: 10, PolicyID: 20}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/bindings/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1BindingsIdGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp BindingResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.BindingId)
	assert.Equal(t, int64(10), resp.SecretId)
	assert.Equal(t, int64(20), resp.PolicyId)
}

func TestBindingAPI_GetByID_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestBindingAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/bindings/", nil)

	api.ApiV1BindingsIdGet(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBindingAPI_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockBindingCase := newTestBindingAPI(ctrl)

	mockBindingCase.EXPECT().
		DeleteBinding(gomock.Any(), "1").
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/bindings/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1BindingsIdDelete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBindingAPI_Delete_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestBindingAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/bindings/", nil)

	api.ApiV1BindingsIdDelete(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBindingAPI_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockBindingCase := newTestBindingAPI(ctrl)

	mockBindingCase.EXPECT().
		CreateBinding(gomock.Any(), gomock.Any()).
		Return(&model.Binding{ID: 1, SecretID: 10, PolicyID: 20, Username: "testuser"}, nil)

	body := `{"secret_id":10,"policy_id":20,"username":"testuser"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/bindings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1BindingsPost(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp BindingResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.BindingId)
	assert.Equal(t, int64(10), resp.SecretId)
	assert.Equal(t, int64(20), resp.PolicyId)
}

func TestBindingAPI_Create_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestBindingAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/bindings", strings.NewReader(`{invalid}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1BindingsPost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBindingAPI_Create_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockBindingCase := newTestBindingAPI(ctrl)

	mockBindingCase.EXPECT().
		CreateBinding(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("create failed"))

	body := `{"secret_id":10,"policy_id":20,"username":"testuser"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/bindings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1BindingsPost(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
