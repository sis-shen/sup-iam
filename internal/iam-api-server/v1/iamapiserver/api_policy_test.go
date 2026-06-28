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

func newTestPolicyAPI(ctrl *gomock.Controller) (*PolicyAPI, *servicemock.MockPolicyCaseInterface) {
	mockPolicyCase := servicemock.NewMockPolicyCaseInterface(ctrl)
	logger := testutil.NewTestLogger()
	api := NewPolicyAPI(mockPolicyCase, logger)
	return api, mockPolicyCase
}

func TestPolicyAPI_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	content := "{}"
	mockPolicyCase.EXPECT().
		GetPolicyList(gomock.Any(), "1", gomock.Any()).
		Return(repository.PageResult[*model.Policy]{
			Items: []*model.Policy{
				{ID: 1, Name: "policy-1", PolicyShadow: &content},
				{ID: 2, Name: "policy-2", PolicyShadow: &content},
			},
			Total: 2,
		}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/policies?page=1&page_size=20", nil)

	api.ApiV1PoliciesGet(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PolicyListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 2, len(resp.Items))
	assert.Equal(t, int64(2), resp.Total)
}

func TestPolicyAPI_List_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestPolicyAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 不设置 UserIDKey
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/policies", nil)

	api.ApiV1PoliciesGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPolicyAPI_List_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	mockPolicyCase.EXPECT().
		GetPolicyList(gomock.Any(), "1", gomock.Any()).
		Return(repository.PageResult[*model.Policy]{}, errors.New("database error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/policies", nil)

	api.ApiV1PoliciesGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPolicyAPI_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	content := "{}"
	mockPolicyCase.EXPECT().
		GetPolicyByID(gomock.Any(), "1").
		Return(&model.Policy{ID: 1, Name: "policy-1", PolicyShadow: &content}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/policies/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1PoliciesIdGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp PolicyResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, "policy-1", resp.Name)
}

func TestPolicyAPI_GetByID_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)
	mockPolicyCase.EXPECT().GetPolicyByID(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/policies/", nil)

	api.ApiV1PoliciesIdGet(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPolicyAPI_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	mockPolicyCase.EXPECT().
		DeletePolicy(gomock.Any(), "1").
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/policies/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1PoliciesIdDelete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPolicyAPI_Delete_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)
	mockPolicyCase.EXPECT().DeletePolicy(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/policies/", nil)

	api.ApiV1PoliciesIdDelete(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPolicyAPI_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	mockPolicyCase.EXPECT().
		CreatePolicy(gomock.Any(), gomock.Any()).
		Return(&model.Policy{ID: 1, Name: "policy-1"}, nil)

	body := `{"name":"policy-1","user_name":"testuser"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/policies", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1PoliciesPost(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPolicyAPI_Create_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)
	mockPolicyCase.EXPECT().CreatePolicy(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/policies", strings.NewReader(`{invalid}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1PoliciesPost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPolicyAPI_Create_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	mockPolicyCase.EXPECT().
		CreatePolicy(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("create failed"))

	body := `{"name":"policy-1","user_name":"testuser"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/policies", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1PoliciesPost(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPolicyAPI_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	desc := "old desc"
	content := "{}"
	extShadow := "{}"
	mockPolicyCase.EXPECT().
		GetPolicyByID(gomock.Any(), "1").
		Return(&model.Policy{
			ID: 1, Name: "policy-1", Description: &desc,
			PolicyShadow: &content, ExtendShadow: &extShadow,
		}, nil)
	mockPolicyCase.EXPECT().
		UpdatePolicy(gomock.Any(), gomock.Any()).
		Return(&model.Policy{ID: 1, Name: "policy-1", PolicyShadow: &content}, nil)

	body := `{"name":"policy-1-updated","description":"new desc"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/policies/1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1PoliciesIdPut(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPolicyAPI_Update_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)
	mockPolicyCase.EXPECT().UpdatePolicy(gomock.Any(), gomock.Any()).Times(0)
	mockPolicyCase.EXPECT().GetPolicyByID(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/policies/", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1PoliciesIdPut(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPolicyAPI_Update_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)
	mockPolicyCase.EXPECT().UpdatePolicy(gomock.Any(), gomock.Any()).Times(0)
	mockPolicyCase.EXPECT().GetPolicyByID(gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/policies/1", strings.NewReader(`{invalid}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1PoliciesIdPut(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPolicyAPI_Update_GetPolicyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	mockPolicyCase.EXPECT().
		GetPolicyByID(gomock.Any(), "1").
		Return(nil, errors.New("not found"))

	body := `{"name":"policy-1"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/policies/1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1PoliciesIdPut(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestPolicyAPI_GetSecrets_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)

	mockPolicyCase.EXPECT().
		GetPolicyBindingSecretList(gomock.Any(), "1", gomock.Any()).
		Return(repository.PageResult[*model.Secret]{
			Items: []*model.Secret{{ID: 1, AccessKey: "ak-001"}},
			Total: 1,
		}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/policies/1/secrets", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1PoliciesIdSecretsGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp PolicyBindingListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, len(resp.Items))
	assert.Equal(t, "ak-001", resp.Items[0].AccessKey)
}

func TestPolicyAPI_GetSecrets_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockPolicyCase := newTestPolicyAPI(ctrl)
	mockPolicyCase.EXPECT().GetPolicyBindingSecretList(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/policies//secrets", nil)

	api.ApiV1PoliciesIdSecretsGet(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
