package service

import (
	"context"
	"errors"
	"testing"
	"time"

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
	mockRepo.EXPECT().
		GetBindingAuditList(gomock.Any(), repository.PageQuery{Limit: 10}).
		Return(repository.PageResult[*model.BindingAudit]{}, nil)
	auditCase := NewAuditCase(mockRepo)

	result, err := auditCase.GetAuditBindingList(context.Background(), repository.PageQuery{Limit: 10})
	assert.NoError(t, err)
	assert.Empty(t, result.Items)
	assert.Equal(t, int64(0), result.Total)
}

func TestAuditCase_GetAuditBindingByID_ReturnsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	now := time.Now()
	mockRepo.EXPECT().
		GetBindingAuditByID(gomock.Any(), "1").
		Return(&model.BindingAudit{ID: 1, SecretID: 100, PolicyID: 200, Username: "user", CreatedAt: now}, nil)
	auditCase := NewAuditCase(mockRepo)

	result, err := auditCase.GetAuditBindingByID(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, uint64(100), result.SecretID)
	assert.Equal(t, uint64(200), result.PolicyID)
}

func TestAuditCase_GetAuditBindingByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	mockRepo.EXPECT().
		GetBindingAuditByID(gomock.Any(), "999").
		Return(nil, errors.New("not found"))
	auditCase := NewAuditCase(mockRepo)

	_, err := auditCase.GetAuditBindingByID(context.Background(), "999")
	assert.Error(t, err)
}

func TestAuditCase_GetAuditPolicyList_ReturnsEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	mockRepo.EXPECT().
		GetPolicyAuditList(gomock.Any(), repository.PageQuery{Limit: 10}).
		Return(repository.PageResult[*model.PolicyAudit]{}, nil)
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
	now := time.Now()
	desc := "policy-audit"
	mockRepo.EXPECT().
		GetPolicyAuditByID(gomock.Any(), "1").
		Return(&model.PolicyAudit{ID: 1, Name: "policy-1", Username: "user", Description: &desc, CreatedAt: now}, nil)
	auditCase := NewAuditCase(mockRepo)

	result, err := auditCase.GetAuditPolicyByID(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), result.ID)
	assert.Equal(t, "policy-1", result.Name)
}

func TestAuditCase_GetAuditPolicyByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repomock.NewMockAuditRepository(ctrl)
	mockRepo.EXPECT().
		GetPolicyAuditByID(gomock.Any(), "999").
		Return(nil, errors.New("not found"))
	auditCase := NewAuditCase(mockRepo)

	_, err := auditCase.GetAuditPolicyByID(context.Background(), "999")
	assert.Error(t, err)
}
