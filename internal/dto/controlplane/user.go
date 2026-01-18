package controlplane

import "github.com/sis-shen/sup-iam/internal/dto/common"

// ----- Create User ----------

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	IsAdmin  bool   `json:"is_admin"`
}

// -------- Update User --------

type UpdateUserPath struct {
	ID string `uri:"id" binding:"required"`
}
type UpdateUserRequest struct {
	Password *string `json:"password,omitempty"`
	Email    *string `json:"email,omitempty"`
	Phone    *string `json:"phone,omitempty"`
	Nickname *string `json:"nickname,omitempty"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
	IsEnable *bool   `json:"is_enable,omitempty"`
}

// -=--- Get User --------
type GetUserPath struct {
	ID string `uri:"id" binding:"required"`
}
type GetUserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	IsAdmin  bool   `json:"is_admin"`
	IsEnable bool   `json:"is_enable"`
}

// ----- Get User Array ---------

type UserListQuery struct {
	common.PageQuery
}

type UserListResponse struct {
	common.PageResponse
	Items []GetUserResponse `json:"items"`
}

// --------- Delete User ---------

type DeleteUserPath struct {
	ID string `uri:"id" binding:"required"`
}
