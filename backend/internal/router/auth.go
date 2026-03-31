package router

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/pkg/response"
	"backend/pkg/utils"
	"log"
	"net/http"
	"strings"

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

func handleEmailLogin(c *gin.Context) {
	platform := strings.ToLower(strings.TrimSpace(c.GetHeader(middleware.CtxPlatform)))
	if platform != middleware.PlatformWeb && platform != middleware.PlatformMobile {
		log.Printf("platform: %s", platform)
		response.JSON(c, http.StatusBadRequest, false, "Invalid platform", nil)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, false, "Invalid request params", nil)
		return
	}

	if req.Email == "" || req.Password == "" {
		log.Print("email or password is empty")
		response.JSON(c, http.StatusBadRequest, false, "Invalid credentials", nil)
		return
	}

	existingUser, err := model.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("GetUserByEmail error: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "login failed, please try again later", nil)
		return
	}
	if existingUser == nil {
		response.JSON(c, http.StatusUnauthorized, false, "login failed, please check your email or password", nil)
		return
	}

	if err := utils.CheckHashedPassword(existingUser.Password, req.Password); err != nil {
		log.Printf("CheckHashedPassword error: %v", err)
		response.JSON(c, http.StatusUnauthorized, false, "login failed, please check you email or password", nil)
		return
	}

	accessToken, err := utils.GenerateJWT(existingUser.ID, existingUser.Name, platform)
	if err != nil {
		log.Printf("GenerateJWT error: %v", err)
		response.JSON(c, http.StatusUnauthorized, false, "login failed, please try again later", nil)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		log.Printf("GenerateRefreshToken error: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "login failed, please try again later", nil)
		return
	}

	if err := model.UpdateUserRefreshToken(existingUser.ID, platform, refreshToken); err != nil {
		log.Printf("UpdateUserRefreshToken error: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "login failed, please try again later", nil)
		return
	}

	response.JSON(c, http.StatusOK, true, "login successful", gin.H{
		"user":          existingUser,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func handleLogout(c *gin.Context) {
	userIDAny, exist := c.Get(middleware.CtxUserID)
	if !exist {
		log.Print("userID not exists")
		response.JSON(c, http.StatusBadRequest, false, "Unauthorized", nil)
		return
	}
	userID, ok := userIDAny.(int64)
	if !ok {
		log.Print("userID type error")
		response.JSON(c, http.StatusBadRequest, false, "Unauthorized", nil)
		return
	}

	platformAny, exist := c.Get(middleware.CtxPlatform)
	if !exist {
		log.Print("platform not exists")
		response.JSON(c, http.StatusBadRequest, false, "Unauthorized", nil)
		return
	}
	platform, ok := platformAny.(string)
	if !ok {
		log.Print("platform type error")
		response.JSON(c, http.StatusBadRequest, false, "Unauthorized", nil)
		return
	}

	if err := model.DeleteUserRefreshToken(userID, platform); err != nil {
		log.Printf("delete token failed, err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "Logout failed", nil)
		return
	}

	response.JSON(c, http.StatusOK, true, "Logged out", nil)
}

func handleRefreshSession(c *gin.Context) {
	platform := strings.ToLower(strings.TrimSpace(c.GetHeader(middleware.CtxPlatform)))
	if platform != middleware.PlatformWeb && platform != middleware.PlatformMobile {
		log.Printf("platform: %s", platform)
		response.JSON(c, http.StatusBadRequest, false, "Invalid platform", nil)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, false, "Invalid request params", nil)
		return
	}

	if req.RefreshToken == "" {
		log.Print("refresh token is empty")
		response.JSON(c, http.StatusBadRequest, false, "Invalid credentials", nil)
		return
	}

	existingUser, err := model.GetUserByRefreshToken(req.RefreshToken, platform)
	if err != nil || existingUser == nil {
		log.Printf("GetUserByRefreshToken error: %v, user: %v", err, existingUser)
		response.JSON(c, http.StatusInternalServerError, false, "refresh session failed", nil)
		return
	}

	accessToken, err := utils.GenerateJWT(existingUser.ID, existingUser.Name, platform)
	if err != nil {
		log.Printf("GenerateJWT error: %v", err)
		response.JSON(c, http.StatusUnauthorized, false, "refresh session failed", nil)
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		log.Printf("GenerateRefreshToken error: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "refresh session failed", nil)
		return
	}

	if err := model.UpdateUserRefreshToken(existingUser.ID, platform, refreshToken); err != nil {
		log.Printf("UpdateUserRefreshToken error: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "refresh session failed", nil)
		return
	}

	response.JSON(c, http.StatusOK, true, "refresh session successful", gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func handleGetCurrentUser(c *gin.Context) {
	platform := strings.ToLower(strings.TrimSpace(c.GetHeader(middleware.CtxPlatform)))
	if platform != middleware.PlatformWeb && platform != middleware.PlatformMobile {
		log.Printf("platform: %s", platform)
		response.JSON(c, http.StatusBadRequest, false, "Invalid platform", nil)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.JSON(c, http.StatusBadRequest, false, "invalid request params", nil)
		return
	}

	if req.RefreshToken == "" {
		log.Println("refresh token is empty")
		response.JSON(c, http.StatusBadRequest, false, "invalid credentials", nil)
		return
	}

	existingUser, err := model.GetUserByRefreshToken(req.RefreshToken, platform)
	if err != nil || existingUser == nil {
		log.Printf("GetUserByRefreshToken failed, err: %v", err)
		response.JSON(c, http.StatusInternalServerError, false, "invalid credentials", nil)
		return
	}

	response.JSON(c, http.StatusOK, true, "get current user successfully", gin.H{
		"user": existingUser,
	})
}
