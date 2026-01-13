package controlplane

import (
	"github.com/sis-shen/sup-iam/internal/dto/common"
	"time"
)

// ------------- Get Binding ------------------

type GetBindingPath struct {
	ID string `uri:"id" binding:"required"`
}

type GetBindingResponse struct {
	BindingID  string    `json:"binding_id"`
	SecretID   string    `json:"secret_id"`
	PolicyID   string    `json:"policy_id"`
	UserName   string    `json:"user_name"`
	CreateTime time.Time `json:"create_time"`
}

// ------------ Delete Binding -----------

type DeleteBindingPath struct {
	ID string `uri:"id" binding:"required"`
}

// ------- Get Binding List Request -----------

type GetBindingListQuery struct {
	common.PageQuery
	UserID string `form:"user_id"`
}

type GetBindingListResponse struct {
	common.PageResponse
	Items []GetBindingResponse `json:"items"`
}

// ------------- Create Binding -------------

type CreateBindingRequest struct {
	SecretID string `json:"secret_id" binding:"required"`
	PolicyID string `json:"policy_id" binding:"required"`
}
