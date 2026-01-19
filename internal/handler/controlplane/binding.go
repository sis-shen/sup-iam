package controlplane

import "github.com/gin-gonic/gin"

// GetBindingHandler 获取指定Binding信息
// @Summary 获取指定Binding信息
// @Description 根据ID获取指定Binding信息
// @Tags binding
// @Produce json
// @Param id path string true "Binding ID"
// @Success 200 {object} GetBindingResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/bindings/{id} [get]
func GetBindingHandler(c *gin.Context) {

}

// GetBindingListHandler 获取Binding列表
// @Summary 获取Binding列表
// @Description 获取指定用户的Binding列表
// @Tags binding
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} GetBindingListResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/bindings [get]
func GetBindingListHandler(c *gin.Context) {

}

// CraeteBindingHandler 创建Binding
// @Summary 创建Binding
// @Description 创建Binding
// @Tags binding
// @Accept json
// @Produce json
// @Param body body CreateBindingRequest true "Binding信息"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/bindings [post]
func CraeteBindingHandler(c *gin.Context) {

}

// DeleteBindingHandler 删除Binding
// @Summary 删除Binding
// @Description 删除Binding
// @Tags binding
// @Produce json
// @Param id path string true "Policy ID"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/bindings/{id} [delete]
func DeleteBindingHandler(c *gin.Context) {

}
