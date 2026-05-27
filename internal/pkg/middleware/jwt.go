package middleware

import (
	"github.com/sis-shen/sup-iam/internal/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware JWT 认证中间件（通用）
func JWTAuthMiddleware(jwtManager *jwt.Manager, skipPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查跳过路径
		if shouldSkip(c.Request.URL.Path, skipPaths) {
			c.Next()
			return
		}

		// 提取并验证 token
		token := jwtManager.ExtractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "missing authentication token",
			})
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 设置用户信息
		c.Set(jwtManager.GetUserIDKey(), claims.UserID)
		c.Next()
	}
}

// OptionalJWTAuthMiddleware 可选认证
func OptionalJWTAuthMiddleware(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := jwtManager.ExtractToken(c)
		if token != "" {
			if claims, err := jwtManager.ValidateToken(token); err == nil {
				c.Set(jwtManager.GetUserIDKey(), claims.UserID)
			}
		}
		c.Next()
	}
}

func shouldSkip(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}
