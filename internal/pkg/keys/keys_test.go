package keys

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewKeys_DefaultValues(t *testing.T) {
	k := NewKeys(0, 0)
	require.Equal(t, 32, k.AccessKeyLength)
	require.Equal(t, 128, k.SecretKeyLength)
}

func TestNewKeys_CustomValues(t *testing.T) {
	// 低于最小值时会被提升到最小值
	k := NewKeys(16, 64)
	require.Equal(t, 32, k.AccessKeyLength)
	require.Equal(t, 128, k.SecretKeyLength)
}

func TestGenerateAccessKey_Length(t *testing.T) {
	k := NewKeys(32, 128)
	ak := k.GenerateAccessKey()
	// 总长度 = 16(timestampHex) + 2*(AccessKeyLength-20)(randomHex) + 4(counterHex)
	// AccessKeyLength=32 => 16 + 2*(32-20) + 4 = 44
	require.Equal(t, 44, len(ak))
}

func TestGenerateAccessKey_LargerLength(t *testing.T) {
	k := NewKeys(64, 128)
	ak := k.GenerateAccessKey()
	// AccessKeyLength=64 => 16 + 2*(64-20) + 4 = 108
	require.Equal(t, 108, len(ak))
}

func TestGenerateAccessKey_HexEncoding(t *testing.T) {
	k := NewKeys(32, 128)
	ak := k.GenerateAccessKey()
	// 验证结果是合法的 hex 字符串
	_, err := hex.DecodeString(ak)
	require.NoError(t, err)
}

func TestGenerateAccessKey_ContainsTimestampAndCounter(t *testing.T) {
	k := NewKeys(32, 128)
	ak := k.GenerateAccessKey()
	// 前16个字符是时间戳 hex，应由 0-9 a-f 组成
	require.True(t, len(ak) >= 20)
	for _, c := range ak[:16] {
		require.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
	}
	// 最后4个字符是计数器 hex
	for _, c := range ak[len(ak)-4:] {
		require.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
	}
}

func TestGenerateAccessKey_Unique(t *testing.T) {
	k := NewKeys(32, 128)
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		ak := k.GenerateAccessKey()
		require.False(t, keys[ak], "duplicate access key generated")
		keys[ak] = true
	}
}

func TestGenerateSecretKey_Length(t *testing.T) {
	k := NewKeys(32, 128)
	sk := k.GenerateSecretKey()
	// 总长度 = 16(timestampHex) + 2*(SecretKeyLength-20)(randomHex) + 4(counterHex)
	// SecretKeyLength=128 => 16 + 2*(128-20) + 4 = 236
	require.Equal(t, 236, len(sk))
}

func TestGenerateSecretKey_LargerLength(t *testing.T) {
	k := NewKeys(32, 256)
	sk := k.GenerateSecretKey()
	// SecretKeyLength=256 => 16 + 2*(256-20) + 4 = 492
	require.Equal(t, 492, len(sk))
}

func TestGenerateSecretKey_Unique(t *testing.T) {
	k := NewKeys(32, 64)
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		sk := k.GenerateSecretKey()
		require.False(t, keys[sk], "duplicate secret key generated")
		keys[sk] = true
	}
}

func TestGenerateSecretKey_HexEncoding(t *testing.T) {
	k := NewKeys(32, 128)
	sk := k.GenerateSecretKey()
	_, err := hex.DecodeString(sk)
	require.NoError(t, err)
}

func TestSignWithKey_Success(t *testing.T) {
	k := NewKeys(32, 128)
	signature, err := k.SignWithKey("my-secret", "hello world")
	require.NoError(t, err)
	require.NotEmpty(t, signature)
	// HMAC-SHA256 输出 32 字节 = 64 hex 字符
	require.Equal(t, 64, len(signature))
}

func TestSignWithKey_EmptySecretKey(t *testing.T) {
	k := NewKeys(32, 128)
	_, err := k.SignWithKey("", "payload")
	require.Error(t, err)
	require.Contains(t, err.Error(), "secret key is empty")
}

func TestSignWithKey_EmptyPayload(t *testing.T) {
	k := NewKeys(32, 128)
	_, err := k.SignWithKey("secret", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "payload is empty")
}

func TestSignWithKey_Deterministic(t *testing.T) {
	k := NewKeys(32, 128)
	sig1, err := k.SignWithKey("secret", "same-payload")
	require.NoError(t, err)
	sig2, err := k.SignWithKey("secret", "same-payload")
	require.NoError(t, err)
	require.Equal(t, sig1, sig2)
}

func TestSignWithKey_DifferentSecrets(t *testing.T) {
	k := NewKeys(32, 128)
	sig1, _ := k.SignWithKey("secret-1", "payload")
	sig2, _ := k.SignWithKey("secret-2", "payload")
	require.NotEqual(t, sig1, sig2)
}

func TestVerifySecretKey_Success(t *testing.T) {
	k := NewKeys(32, 128)
	signature, err := k.SignWithKey("my-secret", "hello world")
	require.NoError(t, err)

	ok, err := k.VerifySecretKey("my-secret", "hello world", signature)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestVerifySecretKey_WrongSecret(t *testing.T) {
	k := NewKeys(32, 128)
	signature, err := k.SignWithKey("correct-secret", "payload")
	require.NoError(t, err)

	ok, err := k.VerifySecretKey("wrong-secret", "payload", signature)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestVerifySecretKey_WrongPayload(t *testing.T) {
	k := NewKeys(32, 128)
	signature, err := k.SignWithKey("secret", "original-payload")
	require.NoError(t, err)

	ok, err := k.VerifySecretKey("secret", "tampered-payload", signature)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestVerifySecretKey_EmptyHashedStr(t *testing.T) {
	k := NewKeys(32, 128)
	ok, err := k.VerifySecretKey("secret", "payload", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "hashedStr key is empty")
	require.False(t, ok)
}

func TestVerifySecretKey_InvalidHashedStr(t *testing.T) {
	k := NewKeys(32, 128)
	ok, err := k.VerifySecretKey("secret", "payload", "not-a-valid-hex-signature")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestGenerateSecretKey_Concurrent(t *testing.T) {
	k := NewKeys(32, 128)
	n := 50
	keys := make(chan string, n)
	for i := 0; i < n; i++ {
		go func() {
			keys <- k.GenerateSecretKey()
		}()
	}

	seen := make(map[string]bool)
	for i := 0; i < n; i++ {
		key := <-keys
		if seen[key] {
			t.Errorf("duplicate secret key generated concurrently: %s", key)
		}
		seen[key] = true
	}
	require.Equal(t, n, len(seen), "all %d keys should be unique", n)
}

func TestGenerateAccessKey_Concurrent(t *testing.T) {
	k := NewKeys(32, 128)
	n := 50
	keys := make(chan string, n)
	for i := 0; i < n; i++ {
		go func() {
			keys <- k.GenerateAccessKey()
		}()
	}

	seen := make(map[string]bool)
	for i := 0; i < n; i++ {
		key := <-keys
		if seen[key] {
			t.Errorf("duplicate access key generated concurrently: %s", key)
		}
		seen[key] = true
	}
	require.Equal(t, n, len(seen), "all %d keys should be unique", n)
}

func TestKeysInterfaceImplementation(t *testing.T) {
	k := NewKeys(32, 128)
	var _ KeysInterface = k
}
