package rpc

import (
	"context"
	"testing"

	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service/mock"
	pbv2 "github.com/sis-shen/sup-iam/internal/pkg/proto/rpc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewAuthQueryHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)
	require.NotNil(t, h)
}

func TestAuthQueryHandler_ImplementsServer(t *testing.T) {
	var _ pbv2.AuthQueryServiceServer = (*AuthQueryHandler)(nil)
}

func TestGetAllSecrets_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetAllSecrets(gomock.Any()).
		Return([]*model.Secret{
			{ID: 1, AccessKey: "AK-test-1", SecretKey: "sk-test-1", Expires: 1700000000},
			{ID: 2, AccessKey: "AK-test-2", SecretKey: "sk-test-2", Expires: 1800000000},
		}, nil)

	resp, err := h.GetAllSecrets(context.Background(), &pbv2.GetAllSecretsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Secrets, 2)

	assert.Equal(t, "1", resp.Secrets[0].SecretId)
	assert.Equal(t, "AK-test-1", resp.Secrets[0].AccessKey)
	assert.Equal(t, "sk-test-1", resp.Secrets[0].SecretKey)
	assert.Equal(t, int64(1700000000), resp.Secrets[0].ExpiresAt)

	assert.Equal(t, "2", resp.Secrets[1].SecretId)
	assert.Equal(t, "AK-test-2", resp.Secrets[1].AccessKey)
}

func TestGetAllSecrets_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetAllSecrets(gomock.Any()).
		Return([]*model.Secret{}, nil)

	resp, err := h.GetAllSecrets(context.Background(), &pbv2.GetAllSecretsRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.Secrets)
}

func TestGetAllSecrets_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetAllSecrets(gomock.Any()).
		Return(nil, assert.AnError)

	resp, err := h.GetAllSecrets(context.Background(), &pbv2.GetAllSecretsRequest{})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetAllPolicies_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	dsl1 := `[["alice", "/api/resource", "GET"]]`
	dsl2 := `[["bob", "/api/other", "POST"]]`

	mockCase.EXPECT().
		GetAllPolicies(gomock.Any()).
		Return(map[string][]*model.Policy{
			"s1": {
				{ID: 100, Username: "alice", PolicyShadow: &dsl1},
				{ID: 101, Username: "alice", PolicyShadow: &dsl2},
			},
			"s2": {
				{ID: 200, Username: "bob", PolicyShadow: strPtr(`[["bob", "/api/admin", "GET"]]`)},
			},
		}, nil)

	resp, err := h.GetAllPolicies(context.Background(), &pbv2.GetAllPoliciesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.PolicyGroups, 2)

	policyCount := 0
	for _, group := range resp.PolicyGroups {
		for _, p := range group.PolicyGroup {
			policyCount++
			switch p.PolicyId {
			case "100":
				assert.Equal(t, "s1", p.SecretId)
				assert.Equal(t, "alice", p.Username)
				assert.Equal(t, dsl1, p.PolicyDsl)
			case "101":
				assert.Equal(t, "s1", p.SecretId)
				assert.Equal(t, "alice", p.Username)
				assert.Equal(t, dsl2, p.PolicyDsl)
			case "200":
				assert.Equal(t, "s2", p.SecretId)
				assert.Equal(t, "bob", p.Username)
			default:
				t.Errorf("unexpected policy id: %s", p.PolicyId)
			}
		}
	}
	assert.Equal(t, 3, policyCount)
}

func TestGetAllPolicies_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetAllPolicies(gomock.Any()).
		Return(map[string][]*model.Policy{}, nil)

	resp, err := h.GetAllPolicies(context.Background(), &pbv2.GetAllPoliciesRequest{})
	require.NoError(t, err)
	assert.Empty(t, resp.PolicyGroups)
}

func TestGetAllPolicies_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	mockCase.EXPECT().
		GetAllPolicies(gomock.Any()).
		Return(nil, assert.AnError)

	resp, err := h.GetAllPolicies(context.Background(), &pbv2.GetAllPoliciesRequest{})
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetAllPolicies_SecretIDInPolicyGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCase := mock.NewMockSecretCaseInterface(ctrl)
	h := NewAuthQueryHandler(mockCase)

	dsl := `[["alice", "/api/test", "GET"]]`
	mockCase.EXPECT().
		GetAllPolicies(gomock.Any()).
		Return(map[string][]*model.Policy{
			"secret-123": {
				{ID: 1, Username: "alice", PolicyShadow: &dsl},
			},
		}, nil)

	resp, err := h.GetAllPolicies(context.Background(), &pbv2.GetAllPoliciesRequest{})
	require.NoError(t, err)
	require.Len(t, resp.PolicyGroups, 1)
	require.Len(t, resp.PolicyGroups[0].PolicyGroup, 1)

	p := resp.PolicyGroups[0].PolicyGroup[0]
	assert.Equal(t, "secret-123", p.SecretId)
	assert.Equal(t, "1", p.PolicyId)
	assert.Equal(t, "alice", p.Username)
	assert.Equal(t, dsl, p.PolicyDsl)
}

func strPtr(s string) *string { return &s }
