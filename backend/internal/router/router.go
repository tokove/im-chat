package router

import (
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CorsMiddleware())
	r.Static("/files", "./files") 

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
		protected.GET("/auth/current-user", handleGetCurrentUser)

		// Users
		protected.GET("/users/:id", handleGetUserByID)

		// Conversations
		protected.GET("/api/conversations/privates/:private_id", handleGetPrivate)
		protected.POST("/api/conversations/privates/create", handleCreatePrivate)
		protected.GET("/api/conversations", handleGetConversations)
		protected.GET("/api/conversations/privates/:private_id/messages", handleGetPrivateMessages)

		// Files
		protected.POST("/api/files/:private_id", handleFileUpload)
	}

	return r
}
