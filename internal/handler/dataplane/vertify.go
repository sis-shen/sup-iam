package dataplane

import "github.com/gin-gonic/gin"

// VerifyHandler 鉴权验证
// @Summary 鉴权验证
// @Description 鉴权验证
// @Tags auth-verify
// @Accept json
// @Produce json
// @Param body body VerifyRequest true "鉴权请求"
// @Success 200 {object} VerifyResponse
// @Failure 401 {object} ErrorResponse "鉴权失败"
// @Failure 403 {object} ErrorResponse "权限不足"
// @Router /auth/v1/verify [post]
func VerifyHandler(c *gin.Context) {

}
