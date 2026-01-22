package main

import (
	"github.com/gin-gonic/gin"
	rt "github.com/sis-shen/sup-iam/internal/router"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title IAM API
// @version 1.0
// @description IAM Control Plane + Auth Plane API
// @BasePath /
func main() {
	r := gin.Default()

	// 注册 API 路由
	rt.RegisterControlPlaneRouter(r)
	rt.RegisterDataPlaneRouter(r)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
