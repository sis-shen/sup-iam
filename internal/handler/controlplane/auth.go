package controlplane

import "github.com/gin-gonic/gin"

// LoginHandler 用户登录
// @Summary 用户登录
// @Description 用户输入用户名/邮箱/手机号和密码登录
// @Tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "登录信息"
// @Success 200 {object} LoginResponse
// @Failure 401 {object} ErrorResponse "登录认证失败"
// @Router /api/v1/auth/login [post]
func LoginHandler(c *gin.Context) {

}

// RegisterHandler 用户注册
// @Summary 用户注册
// @Description 用户输入用户名、邮箱（可选）、手机号（可选）和密码注册
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "注册信息"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse "参数非法"
// @Failure 409 {object} ErrorResponse "用户已存在"
// @Router /api/v1/auth/register [post]
func RegisterHandler(c *gin.Context) {

}

// LogoutHandler 用户退出登录
// @Summary 用户退出登录
// @Description 用户从登录状态退出
// @Tags auth
// @Produce json
// @Success 200 {object} LogoutResponse
// @Failure 401 {object} ErrorResponse "用户未登录"
// @Router /api/v1/auth/logout [post]
func LogoutHandler(c *gin.Context) {

}

// MeHandler 获取用户信息
// @Summary 获取用户信息
// @Description 利用Token获取用户信息
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} MeResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/auth/me [get]
func MeHandler(c *gin.Context) {

}

// RefreshHandler 刷新Token
// @Summary 刷新Token
// @Description 用户使用Refresh Token请求刷新Access Token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body RefreshTokenRequest true "刷新Token请求"
// @Success 200 {object} RefreshResponse
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/auth/refresh [post]
func RefreshHandler(c *gin.Context) {

}

// ChangePasswordHandler 修改密码
// @Summary 修改密码
// @Description 用户上传新的密码用来修改登录代码
// @Tags auth
// @Accept json
// @Produce json
// @Param body body ChangePasswordRequest true "登录信息"
// @Success 200 "OK"
// @Failure 401 {object} ErrorResponse "Token认证失败"
// @Router /api/v1/auth/password/change [post]
func ChangePasswordHandler(c *gin.Context) {

}
