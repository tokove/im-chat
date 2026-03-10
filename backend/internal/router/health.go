package router

import (
	"backend/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleHealthCheckHTTP(c *gin.Context) {
	response.JSON(c, http.StatusOK, true, "API is healthy", nil)
}
