package controlplane

import (
	"github.com/sis-shen/sup-iam/internal/dto/common"
	"time"
)

// -------- Get Secret ----------

type GetSecretPath struct {
	ID string `uri:"id" binding:"required"`
}
type GetSecretResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	AccessKey   string `json:"access_key"`
	Description string `json:"description"`
	Expires     int64  `json:"expires"`
	CreateTime  int64  `json:"create_time"`
	UpdateTime  int64  `json:"update_time"`
}

// --------- Get Secret List ---------

type GetSecretListQuery struct {
	common.PageQuery
}

type GetSecretListResponse struct {
	common.PageResponse
	Items []GetSecretResponse `json:"items"`
}

// ------------- Create Secret -------

type CreateSecretRequest struct {
	Description string `json:"description,omitempty"`
	Expires     int64  `json:"expires,omitempty"`
}
type CreateSecretResponse struct {
	UserID     string    `json:"user_id"`
	AccessKey  string    `json:"access_key"`
	SecretKey  string    `json:"secret_key"`
	Expires    int64     `json:"expires"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// ----------------- Update Secret ------------

type UpdateSecretPath struct {
	ID string `uri:"id" binding:"required"`
}
type UpdateSecretRequest struct {
	Description string `json:"description,omitempty"`
}

// ------------- Delete Secret -------------

type DeleteSecretPath struct {
	ID string `uri:"id" binding:"required"`
}

// ------------ Rotate Secret --------------

type RotateSecretPath struct {
	ID string `uri:"id" binding:"required"`
}

type RotateSecretResponse struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Expires   int64  `json:"expires"`
}

// ----------- Get Secret Binding Policy List -------

type GetSecretBindingPolicyListQuery struct {
	common.PageQuery
	SecretID string `form:"secret_id" binding:"required"`
}

type GetSecretBindingPolicyListResponse struct {
	common.PageResponse
	Items []GetPolicyResponse `json:"items"`
}
