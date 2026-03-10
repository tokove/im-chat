package router

import (
	"backend/internal/model"
	"backend/pkg/response"
	"backend/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleEmailRegister(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, false, "Invalid request params", nil)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		response.JSON(c, http.StatusBadRequest, false, "Invalid credentials", nil)
		return
	}

	existingUser, _ := model.GetUserByEmail(req.Email)
	if existingUser != nil {
		response.JSON(c, http.StatusConflict, false, "Email already in use", nil)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		response.JSON(c, http.StatusInternalServerError, false, "Signup failed, please try again later", nil)
		return
	}

	user, err := model.CreateUserByEmail(req.Name, req.Email, hashedPassword)
	if err != nil {
		response.JSON(c, http.StatusInternalServerError, false, "Signup failed, please try again later", nil)
		return
	}

	response.JSON(c, http.StatusCreated, true, "Signup successful", user)
}
