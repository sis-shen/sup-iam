package jwt

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testConfig() *Config {
	return &Config{
		SecretKey:         "test-secret-key-for-jwt-testing",
		TokenExpireTime:   time.Hour,
		RefreshExpireTime: 24 * time.Hour,
		Issuer:            "test-issuer",
	}
}

func TestNewManager_DefaultValues(t *testing.T) {
	m := NewManager(&Config{SecretKey: "secret"})
	require.Equal(t, 2*time.Hour, m.config.TokenExpireTime)
	require.Equal(t, 7*24*time.Hour, m.config.RefreshExpireTime)
	require.Equal(t, "header:Authorization", m.config.TokenLookup)
	require.Equal(t, "iam-system", m.config.Issuer)
	require.Equal(t, "user_id", m.config.UserIDKey)
}

func TestNewManager_CustomValues(t *testing.T) {
	m := NewManager(testConfig())
	require.Equal(t, time.Hour, m.config.TokenExpireTime)
	require.Equal(t, 24*time.Hour, m.config.RefreshExpireTime)
	require.Equal(t, "test-issuer", m.config.Issuer)
}

func TestGenerateToken_Success(t *testing.T) {
	m := NewManager(testConfig())

	accessToken, refreshToken, err := m.GenerateToken("user-001", "alice")
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, refreshToken)
	// JWT 由三部分组成，各有2个点分隔
	require.Equal(t, 3, len(strings.Split(accessToken, ".")))
	require.Equal(t, 3, len(strings.Split(refreshToken, ".")))
}

func TestGenerateToken_UniqueTokens(t *testing.T) {
	m := NewManager(testConfig())

	tokens := make(map[string]bool)
	for i := 0; i < 10; i++ {
		ak, rk, err := m.GenerateToken("user-001", "alice")
		require.NoError(t, err)
		require.False(t, tokens[ak], "duplicate access token")
		require.False(t, tokens[rk], "duplicate refresh token")
		tokens[ak] = true
		tokens[rk] = true
	}
}

func TestValidateToken_Success(t *testing.T) {
	m := NewManager(testConfig())

	accessToken, _, err := m.GenerateToken("user-001", "alice")
	require.NoError(t, err)

	claims, err := m.ValidateToken(accessToken)
	require.NoError(t, err)
	require.Equal(t, "user-001", claims.UserID)
	require.Equal(t, "alice", claims.Username)
	require.Equal(t, "test-issuer", claims.Issuer)
	require.Equal(t, "user-001", claims.Subject)
	// 验证过期时间在未来
	require.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	m := NewManager(testConfig())
	m2 := NewManager(&Config{
		SecretKey: "different-secret-key",
		Issuer:    "test-issuer",
	})

	accessToken, _, err := m2.GenerateToken("user-001", "alice")
	require.NoError(t, err)

	_, err = m.ValidateToken(accessToken)
	require.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:       "test-secret",
		TokenExpireTime: -time.Hour, // 已过期
		Issuer:          "test-issuer",
	})

	accessToken, _, err := m.GenerateToken("user-001", "alice")
	require.NoError(t, err)

	_, err = m.ValidateToken(accessToken)
	require.Error(t, err)
	require.ErrorContains(t, err, "token is expired")
}

func TestValidateToken_Malformed(t *testing.T) {
	m := NewManager(testConfig())

	_, err := m.ValidateToken("not-a-valid-token")
	require.Error(t, err)
}

func TestValidateToken_WrongIssuer(t *testing.T) {
	m := NewManager(testConfig())
	mDiff := NewManager(&Config{
		SecretKey: "test-secret-key-for-jwt-testing",
		Issuer:    "different-issuer",
	})

	// 用不同签发者生成 token
	accessToken, _, err := mDiff.GenerateToken("user-001", "alice")
	require.NoError(t, err)

	// 用原始 manager 验证（相同密钥但不同签发者）
	_, err = m.ValidateToken(accessToken)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid issuer")
}

func TestValidateRefreshToken_Success(t *testing.T) {
	m := NewManager(testConfig())

	_, refreshToken, err := m.GenerateToken("user-001", "alice")
	require.NoError(t, err)

	claims, err := m.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	require.Equal(t, "test-issuer", claims.Issuer)
	require.Equal(t, "user-001", claims.Subject)
	require.True(t, claims.ExpiresAt.Time.After(time.Now()))
}

func TestValidateRefreshToken_Expired(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:         "test-secret",
		RefreshExpireTime: -time.Hour,
		Issuer:            "test-issuer",
	})

	_, refreshToken, err := m.GenerateToken("user-001", "alice")
	require.NoError(t, err)

	_, err = m.ValidateRefreshToken(refreshToken)
	require.Error(t, err)
}

func TestValidateRefreshToken_Malformed(t *testing.T) {
	m := NewManager(testConfig())

	_, err := m.ValidateRefreshToken("invalid-token")
	require.Error(t, err)
}

func TestExtractToken_FromHeaderBearer(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "header:Authorization",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer my-test-token")

	token := m.ExtractToken(c)
	require.Equal(t, "my-test-token", token)
}

func TestExtractToken_FromHeaderRaw(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "header:Authorization",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "raw-token-value")

	token := m.ExtractToken(c)
	require.Equal(t, "raw-token-value", token)
}

func TestExtractToken_FromHeaderMissing(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "header:Authorization",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	token := m.ExtractToken(c)
	require.Empty(t, token)
}

func TestExtractToken_FromQuery(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "query:token",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?token=query-token-value", nil)

	token := m.ExtractToken(c)
	require.Equal(t, "query-token-value", token)
}

func TestExtractToken_FromQueryMissing(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "query:token",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	token := m.ExtractToken(c)
	require.Empty(t, token)
}

func TestExtractToken_FromCookie(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "cookie:session",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Cookie", "session=cookie-token-value")

	token := m.ExtractToken(c)
	require.Equal(t, "cookie-token-value", token)
}

func TestExtractToken_FromCookieMissing(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "cookie:session",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	token := m.ExtractToken(c)
	require.Empty(t, token)
}

func TestExtractToken_InvalidLookupFormat(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "invalidformat",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer fallback-token")

	token := m.ExtractToken(c)
	require.Equal(t, "fallback-token", token)
}

func TestExtractToken_UnknownSource(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:   "secret",
		TokenLookup: "unknown:token",
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Header.Set("Authorization", "Bearer default-token")

	token := m.ExtractToken(c)
	require.Equal(t, "default-token", token)
}

func TestGetTokenExpireTime(t *testing.T) {
	m := NewManager(&Config{
		SecretKey:       "secret",
		TokenExpireTime: 30 * time.Minute,
	})
	require.Equal(t, 30*time.Minute, m.GetTokenExpireTime())
}

func TestGetUserIDKey(t *testing.T) {
	m := NewManager(&Config{
		SecretKey: "secret",
		UserIDKey: "custom_user_id",
	})
	require.Equal(t, "custom_user_id", m.GetUserIDKey())
}

func TestGenerateAndValidateFullFlow(t *testing.T) {
	m := NewManager(testConfig())

	accessToken, refreshToken, err := m.GenerateToken("user-100", "bob")
	require.NoError(t, err)

	// 验证 access token
	claims, err := m.ValidateToken(accessToken)
	require.NoError(t, err)
	require.Equal(t, "user-100", claims.UserID)
	require.Equal(t, "bob", claims.Username)

	// 验证 refresh token
	refreshClaims, err := m.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	require.Equal(t, "user-100", refreshClaims.Subject)

	// access token 和 refresh token 不同
	require.NotEqual(t, accessToken, refreshToken)
}

func TestTokenHMACAlgorithm(t *testing.T) {
	m := NewManager(testConfig())

	accessToken, _, err := m.GenerateToken("user-001", "alice")
	require.NoError(t, err)

	// 通过 ValidateToken 间接验证算法 (HS256)
	claims, err := m.ValidateToken(accessToken)
	require.NoError(t, err)
	require.NotNil(t, claims)
}

func TestJwtRegisteredClaimsCompatibility(t *testing.T) {
	// 验证 CustomClaims 包含所有必要字段
	claims := &CustomClaims{
		UserID:   "u1",
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:      "jti-123",
			Issuer:  "test",
			Subject: "u1",
		},
	}
	require.Equal(t, "u1", claims.UserID)
	require.Equal(t, "alice", claims.Username)
	require.Equal(t, "jti-123", claims.ID)
	require.Equal(t, "test", claims.Issuer)
	require.Equal(t, "u1", claims.Subject)
}
