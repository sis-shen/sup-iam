package controlplane

import (
	"github.com/sis-shen/sup-iam/internal/dto/common"
	"time"
)

// ----------- Get Policy Audit ------------

type GetPolicyAuditPath struct {
	ID string `uri:"id" binding:"required"`
}

type GetPolicyAuditResponse struct {
	PolicyAuditID string    `json:"policy_audit_id"`
	Name          string    `json:"name"`
	Username      string    `json:"username"`
	Description   string    `json:"description"`
	PolicyShadow  string    `json:"policy_shadow"`
	ExtendShadow  string    `json:"extend_shadow"`
	CreateTime    time.Time `json:"create_time"`
	ActionContent string    `json:"action_content"`
}

// ----------- Get Policy Audit List ----------

type GetPolicyAuditListQuery struct {
	common.PageQuery
	PolicyID string `form:"policy_id,omitempty"`
	UserID   string `form:"user_id,omitempty"`
}

type GetPolicyAuditListResponse struct {
	common.PageResponse
	Items []GetPolicyAuditResponse `json:"items"`
}

// ------ Get Binding Audit-----------

type GetBindingAuditPath struct {
	ID string `uri:"id" binding:"required"`
}
type GetBindingAuditResponse struct {
	BindingAuditID string    `json:"binding_audit_id"`
	SecretID       string    `json:"secret_id"`
	PolicyID       string    `json:"policy_id"`
	Username       string    `json:"username"`
	ActionContent  string    `json:"action_content"`
	CreateTime     time.Time `json:"create_time"`
}

// --------- Get Binding Audit List -------

type GetBindingAuditListQuery struct {
	common.PageQuery
	SecretID string `form:"secret_id,omitempty"`
	PolicyID string `form:"policy_id,omitempty"`
	UserID   string `form:"user_id,omitempty"`
}

type GetBindingAuditListResponse struct {
	common.PageResponse
	Items []GetBindingAuditResponse `json:"items"`
}
