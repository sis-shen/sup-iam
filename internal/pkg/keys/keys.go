package keys

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type KeysInterface interface {
	GenerateSecretKey() string
	GenerateAccessKey() string
	SignWithKey(secretKey string, payload string) (string, error)
	VerifySecretKey(secretKey string, payload string, hashedStr string) (bool, error)
}

type Keys struct {
	AccessKeyLength int `json:"access_key_length"` // AccessKey 长度（位）,至少32位
	SecretKeyLength int `json:"secret_key_length"` // SecretKey 长度（位）,至少128位
}

func NewKeys(accessKeyLength int, secretKeyLength int) *Keys {
	if accessKeyLength < 32 {
		accessKeyLength = 32
	}

	if secretKeyLength < 128 {
		secretKeyLength = 128
	}

	return &Keys{
		AccessKeyLength: accessKeyLength,
		SecretKeyLength: secretKeyLength,
	}

}

var counter uint64

func (k *Keys) GenerateSecretKey() string {
	// 1. 获取纳秒时间戳并转换为16位十六进制
	timestamp := uint64(time.Now().UnixNano())
	timestampHex := fmt.Sprintf("%016x", timestamp)

	// 2. 生成N字节随机数（字节）
	randomBytes := make([]byte, (k.SecretKeyLength-20)/8)
	_, _ = rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	// 3. 获取4位计数器值（0-65535）
	count := atomic.AddUint64(&counter, 1) % 65536
	counterHex := fmt.Sprintf("%04x", count)

	// 4. 组合成N位ID
	return timestampHex + randomHex + counterHex
}

func (k *Keys) GenerateAccessKey() string {
	// 1. 获取纳秒时间戳并转换为16位十六进制
	timestamp := uint64(time.Now().UnixNano())
	timestampHex := fmt.Sprintf("%016x", timestamp)

	// 2. 生成N字节随机数
	randomBytes := make([]byte, (k.AccessKeyLength-20)/8)
	_, _ = rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	// 3. 获取4位计数器值（0-65535）
	count := atomic.AddUint64(&counter, 1) % 65536
	counterHex := fmt.Sprintf("%04x", count)

	// 4. 组合成N位ID
	return timestampHex + randomHex + counterHex
}

func (k *Keys) SignWithKey(secretKey string, payload string) (string, error) {
	if secretKey == "" {
		return "", errors.New("secret key is empty")
	}
	if payload == "" {
		return "", errors.New("payload is empty")
	}

	// 创建 HMAC-SHA256
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(payload))

	// 返回签名（hex 编码）
	return hex.EncodeToString(h.Sum(nil)), nil
}

func (k *Keys) VerifySecretKey(secretKey string, payload string, hashedStr string) (bool, error) {
	if hashedStr == "" {
		return false, errors.New("hashedStr key is empty")
	}
	check, err := k.SignWithKey(secretKey, payload)
	if err != nil {
		return false, err
	}
	if check != hashedStr {
		return false, nil
	}
	return true, nil
}
