package controlplane

import "github.com/gin-gonic/gin"

// GetSecretHandler 获取指定Secret信息
// @Summary 获取指定Secret信息
// @Description 获取指定Secret id的信息
// @Tags secret
// @Produce json
// @Param id path string true "Secret ID"
// @Success 200 {object} GetSecretResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/secrets/{id} [get]
func GetSecretHandler(c *gin.Context) {

}

// GetSecretListHandler 获取指定Secret列表
// @Summary 获取指定Secret列表
// @Description 获取指定User的Secret列表，User由Token指定
// @Tags secret
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} GetSecretListResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/secrets [get]
func GetSecretListHandler(c *gin.Context) {

}

// CreateSecretHandler 创建Secret
// @Summary 创建Secret
// @Description 为指定用户创建Secret
// @Tags secret
// @Accept json
// @Produce json
// @Param body body CreateSecretRequest true "Secret信息"
// @Success 200 {object} CreateSecretResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/secrets [post]
func CreateSecretHandler(c *gin.Context) {

}

// UpdateSecretHandler 更新Secret
// @Summary 更新Secret
// @Description 根据ID更新Secret
// @Tags secret
// @Accept json
// @Param id path string true "Secret ID"
// @Param body body UpdateSecretRequest true "Secret信息"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/secrets/{id} [put]
func UpdateSecretHandler(c *gin.Context) {

}

// DeleteSecretHandler 删除Secret
// @Summary 删除Secret
// @Description 根据ID删除Secret
// @Tags secret
// @Accept json
// @Param id path string true "Secret ID"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/secrets/{id} [delete]
func DeleteSecretHandler(c *gin.Context) {

}

// RotateSecretHandler 轮换Secret
// @Summary 轮换Secret
// @Description 轮换指定Secret的SK，并刷新存活时间
// @Tags secret
// @Accept json
// @Produce json
// @Param id path string true "Secret ID"
// @Success 200 {object} RotateSecretResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/secrets/{id}/rotate [put]
func RotateSecretHandler(c *gin.Context) {

}

// GetSecretBindingPolicyListHandler 获取指定Secret绑定的Policy列表
// @Summary 获取指定Secret绑定的Policy列表
// @Description 获取指定Secret绑定的Policy列表
// @Tags secret
// @Produce json
// @Param id path string true "Secret ID"
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} GetSecretBindingPolicyListResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/secrets/{id}/policies [get]
func GetSecretBindingPolicyListHandler(c *gin.Context) {

}
