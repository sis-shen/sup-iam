package router

import (
	"github.com/gin-gonic/gin"

	dp "github.com/sis-shen/sup-iam/internal/handler/dataplane"
)

func RegisterDataPlaneRouter(r *gin.Engine) {
	// /auth/v1
	v1 := r.Group("/auth/v1")

	v1.POST("/verify", dp.VerifyHandler)
}
