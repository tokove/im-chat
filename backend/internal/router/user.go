package router

import (
	"backend/internal/model"
	"backend/pkg/response"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func handleGetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		log.Println("id parse failed from string")
		response.JSON(c, http.StatusBadRequest, false, "invalid credentails", nil)
		return
	}

	existingUser, err := model.GetUserByID(id)
	if err != nil || existingUser == nil {
		response.JSON(c, http.StatusInternalServerError, false, "invalid credentails", nil)
		return
	}

	response.JSON(c, http.StatusOK, true, "success", gin.H{
		"user": existingUser,
	})
}
