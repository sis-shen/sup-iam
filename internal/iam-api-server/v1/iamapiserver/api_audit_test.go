package iamapiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	servicemock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func newTestAuditAPI(ctrl *gomock.Controller) (*AuditAPI, *servicemock.MockAuditCaseInterface) {
	mockAuditCase := servicemock.NewMockAuditCaseInterface(ctrl)
	api := NewAuditAPI(mockAuditCase)
	return api, mockAuditCase
}

func TestAuditAPI_Bindings_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestAuditAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audits/bindings", nil)

	api.ApiV1AuditsBindingsGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "OK", resp["status"])
}

func TestAuditAPI_BindingsId_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestAuditAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audits/bindings/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1AuditsBindingsIdGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "OK", resp["status"])
}

func TestAuditAPI_Policies_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestAuditAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audits/policies", nil)

	api.ApiV1AuditsPoliciesGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "OK", resp["status"])
}

func TestAuditAPI_PoliciesId_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api, _ := newTestAuditAPI(ctrl)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/audits/policies/1", nil)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	api.ApiV1AuditsPoliciesIdGet(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "OK", resp["status"])
}
