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

func TestAuditCase_GetAuditBindingList_ReturnsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	auditCase := NewAuditCase(mockRepo)

	// AuditCase 当前为 stub，总是返回空结果，不调用 repo
	result, err := auditCase.GetAuditBindingList(context.Background(), repository.PageQuery{Limit: 10})
	assert.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}

func TestAuditCase_GetAuditBindingByID_ReturnsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	auditCase := NewAuditCase(mockRepo)

	result, err := auditCase.GetAuditBindingByID(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, model.BindingAudit{}, result)
}

func TestAuditCase_GetAuditPolicyList_ReturnsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	auditCase := NewAuditCase(mockRepo)

	result, err := auditCase.GetAuditPolicyList(context.Background(), repository.PageQuery{Limit: 10})
	assert.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}

func TestAuditCase_GetAuditPolicyByID_ReturnsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	auditCase := NewAuditCase(mockRepo)

	result, err := auditCase.GetAuditPolicyByID(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, model.PolicyAudit{}, result)
}
