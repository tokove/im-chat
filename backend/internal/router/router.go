package router

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Health Checks
	r.GET("/api/health-check-http", handleHealthCheckHTTP)

	// Auths
	r.POST("/api/auth/register-email", handleEmailRegister)

	return r
}
