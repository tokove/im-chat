package router

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/api/health-check-http", handleHealthCheckHTTP)

	return r
}
