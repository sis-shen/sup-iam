package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Manager JWT 管理器（通用）
type Manager struct {
	config *Config
}

// Config JWT 配置
type Config struct {
	SecretKey         string
	TokenExpireTime   time.Duration
	RefreshExpireTime time.Duration
	Issuer            string //调用方
	TokenLookup       string // header:Authorization, query:token, cookie:token
	UserIDKey         string // 设置进context时userID的key
}

// CustomClaims 自定义 Claims（可根据需要扩展）
type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewManager 创建 JWT 管理器
func NewManager(config *Config) *Manager {
	if config.TokenExpireTime == 0 {
		config.TokenExpireTime = 2 * time.Hour
	}
	if config.RefreshExpireTime == 0 {
		config.RefreshExpireTime = 7 * 24 * time.Hour
	}
	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}
	if config.Issuer == "" {
		config.Issuer = "iam-system"
	}
	if config.UserIDKey == "" {
		config.UserIDKey = "user_id"
	}

	return &Manager{config: config}
}

// GenerateToken 生成 token,返回值为 access_token, refresh_token, error
func (m *Manager) GenerateToken(userID, username string) (string, string, error) {
	// 生成 Access Token
	accessToken, err := m.generateAccessToken(userID, username)
	if err != nil {
		return "", "", err
	}

	// 生成 Refresh Token
	refreshToken, err := m.generateRefreshToken(userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken 生成 Access Token
func (m *Manager) generateAccessToken(userID, username string) (string, error) {
	now := time.Now()

	claims := &CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			Issuer:    m.config.Issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.TokenExpireTime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken 生成 Refresh Token
func (m *Manager) generateRefreshToken(userID string) (string, error) {
	now := time.Now()

	claims := &jwt.RegisteredClaims{
		ID:        uuid.New().String(),
		Issuer:    m.config.Issuer,
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshExpireTime)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证 token，返回 CustomClaims
func (m *Manager) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 token 是否有效
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 获取 claims
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	// 验证签发者（可选，增加安全性）
	if claims.Issuer != m.config.Issuer {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

// ValidateRefreshToken 验证 Refreshtoken
func (m *Manager) ValidateRefreshToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	// 验证签发者
	if claims.Issuer != m.config.Issuer {
		return nil, errors.New("invalid issuer")
	}

	return claims, nil
}

// ExtractToken 从请求中提取 token
func (m *Manager) ExtractToken(c *gin.Context) string {
	// 解析配置 "header:Authorization", "query:token", "cookie:token"
	parts := strings.SplitN(m.config.TokenLookup, ":", 2)
	if len(parts) != 2 {
		// 默认从 header 获取
		return m.extractFromHeader(c, "Authorization")
	}

	source := parts[0]
	name := parts[1]

	switch source {
	case "header":
		return m.extractFromHeader(c, name)
	case "query":
		return c.Query(name)
	case "cookie":
		cookie, err := c.Cookie(name)
		if err == nil {
			return cookie
		}
		return ""
	default:
		return m.extractFromHeader(c, "Authorization")
	}
}

// extractFromHeader 从 header 中提取 Bearer token
func (m *Manager) extractFromHeader(c *gin.Context, headerName string) string {
	authHeader := c.GetHeader(headerName)
	if authHeader == "" {
		return ""
	}

	// Bearer token 格式
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}

	// 如果不是 Bearer 格式，返回原始值
	return authHeader
}

func (m *Manager) GetTokenExpireTime() time.Duration {
	return m.config.TokenExpireTime
}

func (m *Manager) GetUserIDKey() string {
	return m.config.UserIDKey
}
