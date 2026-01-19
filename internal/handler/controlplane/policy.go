package controlplane

import "github.com/gin-gonic/gin"

// GetPolicyHandler 获取指定Policy信息
// @Summary 获取指定Policy信息
// @Description 根据ID获取指定Policy信息
// @Tags policy
// @Produce json
// @Param id path string true "Policy ID"
// @Success 200 {object} GetPolicyResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Failure 404 {object} ErrorResponse "Policy不存在"
// @Router /api/v1/policies/{id} [get]
func GetPolicyHandler(c *gin.Context) {

}

// GetPolicyListHandler 获取Policy列表
// @Summary 获取Policy列表
// @Description 获取指定用户的Policy列表
// @Tags policy
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} GetPolicyListResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/policies [get]
func GetPolicyListHandler(c *gin.Context) {

}

// CreatePolicyHandler 创建Policy
// @Summary 创建Policy
// @Description 根据传入参数创建Policy
// @Tags policy
// @Accept json
// @Param body body CreatePolicyRequest true "Policy信息"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/policies [post]
func CreatePolicyHandler(c *gin.Context) {

}

// UpdatePolicyHandler 更新Policy
// @Summary 更新Policy
// @Description 更新Policy的名称、描述和内容
// @Tags policy
// @Accept json
// @Param id path string true "Policy ID"
// @Param body body UpdatePolicyRequest true "Policy信息"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/policies/{id} [put]
func UpdatePolicyHandler(c *gin.Context) {

}

// DeletePolicyHandler 删除Policy
// @Summary 删除Policy
// @Description 删除指定ID的Policy
// @Tags policy
// @Param id path string true "Policy ID"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/policies/{id} [delete]
func DeletePolicyHandler(c *gin.Context) {

}

// GetPolicyBindingListHandler 获取Policy绑定的Secret列表
// @Summary 获取Policy绑定的Secret列表
// @Description 获取Policy绑定的Secret列表
// @Tags policy
// @Produce json
// @Param id path string true "Policy ID"
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} GetPolicyBindingListResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/policies/{id}/secrets [get]
func GetPolicyBindingListHandler(c *gin.Context) {

}
