package controlplane

import "github.com/gin-gonic/gin"

// CreateUserHandler 创建用户
// @Summary 创建用户
// @Description 管理员调用接口直接创建用户
// @Tags user
// @Accept json
// @Produce json
// @Param body body CreateUserRequest true "创建用户信息"
// @Success 200 "OK"
// @Failure 403 {object} ErrorResponse "没有admin权限"
// @Router /api/v1/users [post]
func CreateUserHandler(c *gin.Context) {

}

// UpdateUserHandler 更新用户字段
// @Summary 更新用户字段
// @Description 更新用户的昵称、手机号、邮箱、管理员权限、是否启用等字段
// @Tags user
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param body body UpdateUserRequest true "User更新信息"
// @Success 200 "OK"
// @Failure 403 {object} ErrorResponse "没有admin权限"
// @Router /api/v1/users/{id} [put]
func UpdateUserHandler(c *gin.Context) {

}

// DeleteUserHandler 删除用户
// @Summary 删除用户
// @Description 删除指定id的用户
// @Tags user
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 "OK"
// @Failure 403 {object} ErrorResponse "没有admin权限"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Router /api/v1/users/{id} [delete]
func DeleteUserHandler(c *gin.Context) {

}

// GetUserHandler 获取指定用户信息
// @Summary 获取指定用户信息
// @Description 获取指定用户id的信息
// @Tags user
// @Produce json
// @Success 200 {object} GetUserResponse
// @Failure 403 {object} ErrorResponse "没有admin权限"
// @Failure 404 {object} ErrorResponse "用户不存在"
// @Router /api/v1/users/{id} [get]
func GetUserHandler(c *gin.Context) {

}

// GetUserListHandler 获取用户列表
// @Summary 获取用户列表
// @Description 按页查询用户列表
// @Tags user
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} GetUserListResponse
// @Failure 403 {object} ErrorResponse "没有admin权限"
// @Router /api/v1/users [get]
func GetUserListHandler(c *gin.Context) {

}
