package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"io/ioutil"
)

type MockConfig struct {
	Users []model.User `json:"users"`
	Delay int          `json:"delay_ms"` // 模拟延迟
}

func main() {
	var (
		port   = flag.Int("port", 8080, "Server port")
		config = flag.String("config", "", "Mock config file path")
		delay  = flag.Int("delay", 0, "Response delay in milliseconds")
	)
	flag.Parse()

	// 加载配置
	var mockData MockConfig
	if *config != "" {
		data, _ := ioutil.ReadFile(*config)
		json.Unmarshal(data, &mockData)
	}

	r := gin.Default()

	// 添加延迟中间件
	if *delay > 0 {
		r.Use(func(c *gin.Context) {
			// time.Sleep(time.Duration(*delay) * time.Millisecond)
		})
	}

	// 动态 Mock 端点（可以动态修改响应）
	r.GET("/mock/users", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		size := c.DefaultQuery("size", "20")

		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "success",
			"data": gin.H{
				"items": mockData.Users,
				"page":  page,
				"size":  size,
				"total": len(mockData.Users),
			},
		})
	})

	// 配置端点（动态修改 Mock 行为）
	r.POST("/mock/config", func(c *gin.Context) {
		var newConfig MockConfig
		if err := c.BindJSON(&newConfig); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		mockData = newConfig
		c.JSON(200, gin.H{"message": "config updated"})
	})

	fmt.Printf("Mock Server running on http://localhost:%d\n", *port)
	fmt.Println("Postman测试地址: http://localhost:8080/mock/users?page=1&size=20")

	r.Run(fmt.Sprintf(":%d", *port))
}
