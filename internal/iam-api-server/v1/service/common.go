package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"strconv"
	"sync/atomic"
	"time"
)

func ParseListQuery(c *gin.Context) (repository.PageQuery, error) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil {
		limit = 10
	}
	cursor := c.Query("next")

	order := repository.Order(c.Query("order"))
	if order != repository.OrderAsc && order != repository.OrderDesc {
		order = repository.OrderAsc
	}

	return repository.PageQuery{Limit: limit, Cursor: cursor, Order: order}, nil
}

func GetPage(c *gin.Context) int32 {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}

	return int32(page)
}

func GetPageSize(c *gin.Context) int32 {
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "1"))
	if err != nil {
		pageSize = 1
	}

	return int32(pageSize)
}

func DerefSlice[T any](list []*T) []T {
	result := make([]T, len(list))
	for i, v := range list {
		result[i] = *v
	}
	return result
}

// 全局计数器，用于防止同一纳秒内的ID冲突
var counter uint64

// GenerateInstanceID 生成32位跨域唯一ID
// 格式：时间戳(16位) + 随机数(12位) + 计数器(4位)
// 总共32位十六进制字符
func GenerateInstanceID() string {
	// 1. 获取纳秒时间戳并转换为16位十六进制
	timestamp := uint64(time.Now().UnixNano())
	timestampHex := fmt.Sprintf("%016x", timestamp)

	// 2. 生成12位随机数（6字节）
	randomBytes := make([]byte, 6)
	_, _ = rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)

	// 3. 获取4位计数器值（0-65535）
	count := atomic.AddUint64(&counter, 1) % 65536
	counterHex := fmt.Sprintf("%04x", count)

	// 4. 组合成32位ID
	return timestampHex + randomHex + counterHex
}
