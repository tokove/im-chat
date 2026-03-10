package response

import "github.com/gin-gonic/gin"

type apiResponse struct {
	Status  int    `json:"status"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func JSON(c *gin.Context, status int, success bool, message string, data any) {
	c.JSON(status, apiResponse{
		Status:  status,
		Success: success,
		Message: message,
		Data:    data,
	})
}
