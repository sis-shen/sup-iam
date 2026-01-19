package controlplane

import "github.com/gin-gonic/gin"

// GetPolicyAuditHandler 获取Policy Audit
// @Summary 获取Policy Audit
// @Description 获取指定ID的Policy Audit
// @Tags auth
// @Produce json
// @Param id path string true "Policy Audit ID"
// @Success 200 {object} GetPolicyAuditResponse
// @Failure 401 {object} ErrorResponse "没有admin权限"
// @Router /api/v1/audits/policies/{id} [get]
func GetPolicyAuditHandler(c *gin.Context) {

}

// GetPolicyAuditListHandler 获取Policy Audit
// @Summary 获取Policy Audit
// @Description 获取Policy Audit列表，允许根据SecretID过滤，PolicyID过滤，UserID过滤
// @Tags auth
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "页大小"
// @Success 200 {object} GetPolicyAuditListResponse
// @Failure 401 {object} ErrorResponse "没有admin权限"
// @Router /api/v1/audits/policies [get]
func GetPolicyAuditListHandler(c *gin.Context) {

}

// GetBindingAuditHandler 获取Binding Audit
// @Summary 获取获取Binding Audit
// @Description 获取指定ID的获取Binding Audit
// @Tags auth
// @Produce json
// @Param id path string true "Binding Audit ID"
// @Success 200 {object} GetBindingAuditListQuery
// @Failure 401 {object} ErrorResponse "没有admin权限"
// @Router /api/v1/audits/bindings/{id} [get]
func GetBindingAuditHandler(c *gin.Context) {

}

// GetBindingAuditListHandler 获取Binding Audit列表
// @Summary 获取获取Binding Audit列表
// @Description 获取获取Binding ,允许根据SecretID过滤，PolicyID过滤，UserID过滤
// @Tags auth
// @Produce json
// @Param id path string true "Binding Audit ID"
// @Success 200 {object} GetBindingAuditListResponse
// @Failure 401 {object} ErrorResponse "没有admin权限"
// @Router /api/v1/audits/bindings/{id} [get]
func GetBindingAuditListHandler(c *gin.Context) {

}
