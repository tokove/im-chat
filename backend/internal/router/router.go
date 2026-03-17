package router

import (
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CorsMiddleware())

	public := r.Group("/api")
	{
		// Health Checks
		public.GET("/health-check-http", handleHealthCheckHTTP)

		// Auths
		public.POST("/auth/register-email", handleEmailRegister)
		public.POST("/auth/login-email", handleEmailLogin)
		public.POST("/auth/refresh-session", handleRefreshSession)
	}

	protected := r.Group("/api")
	protected.Use(middleware.AuthenticateMiddleware())
	{
		// Auths
		protected.POST("/auth/logout", handleLogout)
	}

	return r
}
