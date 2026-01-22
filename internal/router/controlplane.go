package router

import (
	"github.com/gin-gonic/gin"

	cp "github.com/sis-shen/sup-iam/internal/handler/controlplane"
)

func RegisterControlPlaneRouter(r *gin.Engine) {
	// /api/v1
	v1 := r.Group("/api/v1")

	// =========== AUTH =========
	auth := v1.Group("/auth")
	{
		auth.POST("/login", cp.LoginHandler)
		auth.POST("register", cp.RegisterHandler)
		auth.POST("logout", cp.LogoutHandler)
		auth.GET("/me", cp.MeHandler)
		auth.POST("/refresh", cp.RefreshHandler)
		auth.PUT("/password", cp.ChangePasswordHandler)
	}

	users := v1.Group("/users")
	{
		users.GET("", cp.GetUserListHandler)
		users.POST("", cp.CreateUserHandler)
		users.GET("/:id", cp.GetUserHandler)
		users.PUT("/:id", cp.UpdateUserHandler)
		users.DELETE("/:id", cp.DeleteUserHandler)
	}
	// ========== Secret ==========
	secrets := v1.Group("/secrets")
	{
		secrets.GET("", cp.GetSecretListHandler)
		secrets.POST("", cp.CreateSecretHandler)

		secrets.GET("/:id", cp.GetSecretHandler)
		secrets.PUT("/:id", cp.UpdateSecretHandler)
		secrets.DELETE("/:id", cp.DeleteSecretHandler)

		secrets.PUT("/:id/rotate", cp.RotateSecretHandler)
		secrets.GET("/:id/policies", cp.GetSecretBindingPolicyListHandler)
	}

	// ========== Policy ==========
	policies := v1.Group("/policies")
	{
		policies.GET("", cp.GetPolicyListHandler)
		policies.POST("", cp.CreatePolicyHandler)

		policies.GET("/:id", cp.GetPolicyHandler)
		policies.PUT("/:id", cp.UpdatePolicyHandler)
		policies.DELETE("/:id", cp.DeletePolicyHandler)

		policies.GET("/:id/secrets", cp.GetPolicyBindingListHandler)
	}

	// ========== Binding ==========
	bindings := v1.Group("/bindings")
	{
		bindings.GET("", cp.GetBindingListHandler)
		bindings.POST("", cp.CraeteBindingHandler)

		bindings.GET("/:id", cp.GetBindingHandler)
		bindings.DELETE("/:id", cp.DeleteBindingHandler)
	}

	// ========== Audit ==========
	audits := v1.Group("/audits")
	{
		policyAudits := audits.Group("/policies")
		{
			policyAudits.GET("", cp.GetPolicyAuditListHandler)
			policyAudits.GET("/:id", cp.GetPolicyAuditHandler)
		}

		bindingAudits := audits.Group("/bindings")
		{
			bindingAudits.GET("", cp.GetBindingAuditListHandler)
			bindingAudits.GET("/:id", cp.GetBindingAuditHandler)
		}
	}
}
