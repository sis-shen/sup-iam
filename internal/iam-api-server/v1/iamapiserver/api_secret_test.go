package iamapiserver

import (
	"encoding/json"
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

func newTestSecretAPI(ctrl *gomock.Controller) (*SecretAPI, *servicemock.MockSecretCaseInterface) {
	mockSecretCase := servicemock.NewMockSecretCaseInterface(ctrl)
	logger := testutil.NewTestLogger()
	api := NewSecretAPI(mockSecretCase, logger)
	return api, mockSecretCase
}

func TestSecretAPI_List_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockSecretCase := newTestSecretAPI(ctrl)

	mockSecretCase.EXPECT().
		GetSecretList(gomock.Any(), "1", gomock.Any()).
		Return(repository.PageResult[*model.Secret]{
			Items: []*model.Secret{{ID: 1, AccessKey: "ak-001"}},
			Total: 1,
		}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)

	api.ApiV1SecretsGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp SecretListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, len(resp.Items))
}

func TestSecretAPI_List_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 不设置 UserIDKey
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets", nil)

	api.ApiV1SecretsGet(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSecretAPI_GetByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockSecretCase := newTestSecretAPI(ctrl)

	mockSecretCase.EXPECT().
		GetSecretByID(gomock.Any(), "1").
		Return(&model.Secret{ID: 1, AccessKey: "ak-001"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1SecretsIdGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp SecretResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, "ak-001", resp.AccessKey)
}

func TestSecretAPI_GetByID_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets/", nil)

	api.ApiV1SecretsIdGet(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecretAPI_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockSecretCase := newTestSecretAPI(ctrl)

	mockSecretCase.EXPECT().
		DeleteSecret(gomock.Any(), "1").
		Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1SecretsIdDelete(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecretAPI_Delete_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/secrets/", nil)

	api.ApiV1SecretsIdDelete(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecretAPI_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockSecretCase := newTestSecretAPI(ctrl)

	mockSecretCase.EXPECT().
		GenerateAccessKey().
		Return("ak-new")
	mockSecretCase.EXPECT().
		GenerateSecretKey().
		Return("sk-new")
	mockSecretCase.EXPECT().
		CreateSecret(gomock.Any(), gomock.Any()).
		Return(&model.Secret{ID: 1, AccessKey: "ak-new", SecretKey: "sk-new"}, nil)

	body := `{"expires":3600,"user_name":"testuser"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/secrets", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1SecretsPost(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp CreateSecretResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, "ak-new", resp.AccessKey)
}

func TestSecretAPI_Create_NoUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	body := `{"expires":3600}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// 不设置 UserIDKey
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/secrets", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1SecretsPost(c)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSecretAPI_Create_MissingExpires(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(UserIDKey, "1")
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/secrets", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1SecretsPost(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecretAPI_Rotate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockSecretCase := newTestSecretAPI(ctrl)

	mockSecretCase.EXPECT().
		RotateSecret(gomock.Any(), "1").
		Return(&model.Secret{ID: 1, AccessKey: "ak-001", SecretKey: "new-sk"}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/secrets/1/rotate", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1SecretsIdRotatePut(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp RotateSecretResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, int64(1), resp.Id)
	assert.Equal(t, "new-sk", resp.SecretKey)
}

func TestSecretAPI_Rotate_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/secrets//rotate", nil)

	api.ApiV1SecretsIdRotatePut(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecretAPI_Update_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockSecretCase := newTestSecretAPI(ctrl)

	desc := "old desc"
	extShadow := "{}"
	expires := int64(3600)
	mockSecretCase.EXPECT().
		GetSecretByID(gomock.Any(), "1").
		Return(&model.Secret{ID: 1, Description: &desc, Expires: expires, ExtendShadow: &extShadow}, nil)
	mockSecretCase.EXPECT().
		UpdateSecret(gomock.Any(), gomock.Any()).
		Return(&model.Secret{ID: 1, Expires: expires}, nil)

	body := `{"description":"new desc","expires":3600}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/secrets/1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1SecretsIdPut(c)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecretAPI_Update_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/secrets/", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.ApiV1SecretsIdPut(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecretAPI_GetPolicies_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, mockSecretCase := newTestSecretAPI(ctrl)

	content := "{}"
	mockSecretCase.EXPECT().
		GetSecretBindingPolicy(gomock.Any(), "1", gomock.Any()).
		Return(repository.PageResult[*model.Policy]{
			Items: []*model.Policy{{ID: 1, Name: "policy-1", PolicyShadow: &content}},
			Total: 1,
		}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets/1/policies", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1SecretsIdPoliciesGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp SecretPolicyListResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 1, len(resp.Items))
	assert.Equal(t, "policy-1", resp.Items[0].Name)
}

func TestSecretAPI_GetPolicies_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestSecretAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/secrets//policies", nil)

	api.ApiV1SecretsIdPoliciesGet(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
