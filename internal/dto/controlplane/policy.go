package controlplane

import (
	"github.com/sis-shen/sup-iam/internal/dto/common"
	"time"
)

// --------- Get Policy ---------

type GetPolicyPath struct {
	ID string `uri:"id" binding:"required"`
}

type GetPolicyResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Username    string    `json:"username"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// ------- Get Policy List -----

type GetPolicyListQuery struct {
	common.PageQuery
}

type GetPolicyListResponse struct {
	common.PageResponse
	Items []GetPolicyResponse `json:"items"`
}

// -------- Create Policy ----------

type CreatePolicyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// ----------- Update Policy ----------

type UpdatePolicyPath struct {
	ID string `uri:"id" binding:"required"`
}
type UpdatePolicyRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
}

// ---------- Delete Policy -------

type DeletePolicyPath struct {
	ID string `uri:"id" binding:"required"`
}
type DeletePolicyResponse struct {
	Success bool `json:"success"`
}

// ------------ Get Policy Binding Secret List -------

type GetPolicyBindingListQuery struct {
	common.PageQuery
	PolicyID string `form:"policy_id" binding:"required"`
}

type GetPolicyBindingListResponse struct {
	common.PageResponse
	Items []GetSecretResponse `json:"items"`
}
