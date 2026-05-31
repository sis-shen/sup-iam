package rpc

import (
	"context"
	"errors"
	"testing"
	"time"

	pbv1 "github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v1"

	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	servicemock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewAuthQueryHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)
	assert.NotNil(t, h)
	assert.Equal(t, mockCase, h.secretCase)
}

// ---------------------------------------------------------------------------
// GetSecretByAK 测试
// ---------------------------------------------------------------------------

func TestGetSecretByAK_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	now := time.Now()
	secret := &model.Secret{
		ID:        100,
		SecretKey: "sk-001",
		AccessKey: "ak-001",
		Expires:   3600,
		UpdatedAt: now,
	}

	mockCase.EXPECT().
		GetSecretByAK(gomock.Any(), "ak-001").
		Return(secret, nil)

	resp, err := h.GetSecretByAK(context.Background(), &pbv1.GetSecretByAKRequest{AccessKey: "ak-001"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Secret)
	assert.Equal(t, "100", resp.Secret.SecretId)
	assert.Equal(t, "sk-001", resp.Secret.SecretKey)
	assert.Equal(t, "ak-001", resp.Secret.AccessKey)
	// ExpiresAt = UpdatedAt + Expires seconds
	expectedExpires := now.Add(3600 * time.Second).Unix()
	assert.Equal(t, expectedExpires, resp.Secret.ExpiresAt)
}

func TestGetSecretByAK_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetSecretByAK(gomock.Any(), "ak-notfound").
		Return(nil, repository.ErrNotFound)

	resp, err := h.GetSecretByAK(context.Background(), &pbv1.GetSecretByAKRequest{AccessKey: "ak-notfound"})
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetSecretByAK_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetSecretByAK(gomock.Any(), "ak-error").
		Return(nil, errors.New("db error"))

	resp, err := h.GetSecretByAK(context.Background(), &pbv1.GetSecretByAKRequest{AccessKey: "ak-error"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db error")
	assert.Nil(t, resp)
}

// ---------------------------------------------------------------------------
// GetPolicyBySecretID 测试
// ---------------------------------------------------------------------------

func policyShadow(s string) *string {
	return &s
}

func TestGetPolicyBySecretID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	dsl := `[["alice", "/api/resource", "GET"]]`
	policies := []*model.Policy{
		{ID: 10, Username: "alice", PolicyShadow: policyShadow(dsl)},
		{ID: 20, Username: "bob", PolicyShadow: policyShadow(`[["bob", "/api/other", "POST"]]`)},
	}

	mockCase.EXPECT().
		GetSecretBindingPolicy(gomock.Any(), "secret-1", gomock.Any()).
		Return(repository.PageResult[*model.Policy]{Items: policies, Total: 2}, nil)

	resp, err := h.GetPolicyBySecretID(context.Background(), &pbv1.GetPolicyBySecretIDRequest{SecretId: "secret-1"})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.PolicyList, 2)
	assert.Equal(t, "10", resp.PolicyList[0].PolicyId)
	assert.Equal(t, "alice", resp.PolicyList[0].Username)
	assert.Equal(t, dsl, resp.PolicyList[0].PolicyDsl)
	assert.Equal(t, "20", resp.PolicyList[1].PolicyId)
}

func TestGetPolicyBySecretID_SkipsNilPolicyShadow(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	policies := []*model.Policy{
		{ID: 10, Username: "alice", PolicyShadow: nil}, // 应被跳过
		{ID: 20, Username: "bob", PolicyShadow: policyShadow(`[["bob", "/api", "GET"]]`)},
	}

	mockCase.EXPECT().
		GetSecretBindingPolicy(gomock.Any(), "secret-2", gomock.Any()).
		Return(repository.PageResult[*model.Policy]{Items: policies, Total: 2}, nil)

	resp, err := h.GetPolicyBySecretID(context.Background(), &pbv1.GetPolicyBySecretIDRequest{SecretId: "secret-2"})
	assert.NoError(t, err)
	assert.Len(t, resp.PolicyList, 1)
	assert.Equal(t, "20", resp.PolicyList[0].PolicyId)
}

func TestGetPolicyBySecretID_EmptyList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetSecretBindingPolicy(gomock.Any(), "secret-empty", gomock.Any()).
		Return(repository.PageResult[*model.Policy]{Items: []*model.Policy{}, Total: 0}, nil)

	resp, err := h.GetPolicyBySecretID(context.Background(), &pbv1.GetPolicyBySecretIDRequest{SecretId: "secret-empty"})
	assert.NoError(t, err)
	assert.Empty(t, resp.PolicyList)
}

func TestGetPolicyBySecretID_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCase := servicemock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetSecretBindingPolicy(gomock.Any(), "secret-error", gomock.Any()).
		Return(repository.PageResult[*model.Policy]{}, errors.New("db error"))

	resp, err := h.GetPolicyBySecretID(context.Background(), &pbv1.GetPolicyBySecretIDRequest{SecretId: "secret-error"})
	assert.Error(t, err)
	assert.Nil(t, resp)
}
