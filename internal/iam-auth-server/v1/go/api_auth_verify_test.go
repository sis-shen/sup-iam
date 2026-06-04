package iamauthserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	model "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/model"
	servicemock "github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/service/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func init() {
	gin.SetMode(gin.TestMode)
	log.Init(log.NewOptions())
}

func newTestAuthVerifyAPI(t *testing.T, mockSvc *servicemock.MockAuthCaseInterface) *AuthVerifyAPI {
	t.Helper()
	return NewAuthVerifyAPI(mockSvc, log.WithName("test"))
}

func newVerifyRequestBody(t *testing.T, req *VerifyRequest) *strings.Reader {
	t.Helper()
	data, err := json.Marshal(req)
	require.NoError(t, err)
	return strings.NewReader(string(data))
}

func TestNewAuthVerifyAPI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := NewAuthVerifyAPI(mockSvc, log.WithName("test"))
	assert.NotNil(t, api)
	assert.Equal(t, mockSvc, api.service)
}

func TestVerifyRequest_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := newTestAuthVerifyAPI(t, mockSvc)

	req := &VerifyRequest{
		AccessKey:   "ak-001",
		Method:      "GET",
		Path:        "/api/resource",
		ContentHash: "hash123",
		Timestamp:   1700000000,
		Signature:   "valid-sig",
		Username:    "alice",
	}

	canonical := "ak-001\nGET\n/api/resource\nhash123\n1700000000"

	gomock.InOrder(
		mockSvc.EXPECT().BuildCanonicalString("ak-001", "GET", "/api/resource", "hash123", "1700000000").Return(canonical),
		mockSvc.EXPECT().VerifySecretKey("ak-001", canonical, "valid-sig").Return(true, &model.CachedSecret{ID: "100", SecretKey: "sk-001", AccessKey: "ak-001"}, nil),
		mockSvc.EXPECT().Authorize("alice", "ak-001", "/api/resource", "GET").Return(true, []string{"10"}, nil),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/v1/verify", newVerifyRequestBody(t, req))
	c.Request.Header.Set("Content-Type", "application/json")

	api.VerifyRequest(c)

	assert.Equal(t, 200, w.Code)
	var resp VerifyResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Allowed)
	assert.Equal(t, "allowed", *resp.Reason)
	assert.Equal(t, []string{"10"}, resp.MatchedPolicies)
}

func TestVerifyRequest_BindError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := newTestAuthVerifyAPI(t, mockSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/v1/verify", strings.NewReader(`{invalid json`))
	c.Request.Header.Set("Content-Type", "application/json")

	api.VerifyRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errResp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "Parse request body failed", errResp.Error)
}

func TestVerifyRequest_VerifySecretError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := newTestAuthVerifyAPI(t, mockSvc)

	req := &VerifyRequest{
		AccessKey: "ak-001", Method: "GET", Path: "/api/resource",
		ContentHash: "hash", Timestamp: 1700000000, Signature: "sig",
	}

	mockSvc.EXPECT().BuildCanonicalString(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("canonical")
	mockSvc.EXPECT().VerifySecretKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/v1/verify", newVerifyRequestBody(t, req))
	c.Request.Header.Set("Content-Type", "application/json")

	api.VerifyRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errResp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "verify secret error", errResp.Error)
}

func TestVerifyRequest_VerifySecretFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := newTestAuthVerifyAPI(t, mockSvc)

	req := &VerifyRequest{
		AccessKey: "ak-001", Method: "GET", Path: "/api/resource",
		ContentHash: "hash", Timestamp: 1700000000, Signature: "bad-sig",
	}

	secret := &model.CachedSecret{ID: "100", SecretKey: "sk", AccessKey: "ak-001"}

	mockSvc.EXPECT().BuildCanonicalString(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("canonical")
	mockSvc.EXPECT().VerifySecretKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, secret, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/v1/verify", newVerifyRequestBody(t, req))
	c.Request.Header.Set("Content-Type", "application/json")

	api.VerifyRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyRequest_AccessKeyMismatch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := newTestAuthVerifyAPI(t, mockSvc)

	req := &VerifyRequest{
		AccessKey: "ak-001", Method: "GET", Path: "/api/resource",
		ContentHash: "hash", Timestamp: 1700000000, Signature: "sig",
	}

	secret := &model.CachedSecret{ID: "100", SecretKey: "sk", AccessKey: "ak-different"}

	mockSvc.EXPECT().BuildCanonicalString(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("canonical")
	mockSvc.EXPECT().VerifySecretKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, secret, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/v1/verify", newVerifyRequestBody(t, req))
	c.Request.Header.Set("Content-Type", "application/json")

	api.VerifyRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyRequest_AuthorizeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := newTestAuthVerifyAPI(t, mockSvc)

	req := &VerifyRequest{
		AccessKey: "ak-001", Method: "GET", Path: "/api/resource",
		ContentHash: "hash", Timestamp: 1700000000, Signature: "sig",
	}

	secret := &model.CachedSecret{ID: "100", SecretKey: "sk", AccessKey: "ak-001"}

	gomock.InOrder(
		mockSvc.EXPECT().BuildCanonicalString(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("canonical"),
		mockSvc.EXPECT().VerifySecretKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, secret, nil),
		mockSvc.EXPECT().Authorize("", "ak-001", "/api/resource", "GET").Return(false, nil, assert.AnError),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/v1/verify", newVerifyRequestBody(t, req))
	c.Request.Header.Set("Content-Type", "application/json")

	api.VerifyRequest(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVerifyRequest_Denied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := servicemock.NewMockAuthCaseInterface(ctrl)
	api := newTestAuthVerifyAPI(t, mockSvc)

	req := &VerifyRequest{
		AccessKey: "ak-001", Method: "GET", Path: "/api/resource",
		ContentHash: "hash", Timestamp: 1700000000, Signature: "sig",
	}

	secret := &model.CachedSecret{ID: "100", SecretKey: "sk", AccessKey: "ak-001"}

	gomock.InOrder(
		mockSvc.EXPECT().BuildCanonicalString(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("canonical"),
		mockSvc.EXPECT().VerifySecretKey(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, secret, nil),
		mockSvc.EXPECT().Authorize("", "ak-001", "/api/resource", "GET").Return(false, nil, nil),
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/v1/verify", newVerifyRequestBody(t, req))
	c.Request.Header.Set("Content-Type", "application/json")

	api.VerifyRequest(c)

	assert.Equal(t, 403, w.Code)
	var resp VerifyResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Allowed)
	assert.Equal(t, DenyReason, *resp.Reason)
}

func TestDefaultHandleFunc(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	DefaultHandleFunc(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
	assert.Equal(t, "501 not implemented", w.Body.String())
}
