package middleware

import (
	"backend/pkg/response"
	"backend/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CtxUserID          string = "userID"
	CtxUserDisplayName string = "name"
	CtxPlatform        string = "X-Platform"
	CtxAuthorization   string = "Authorization"
	PlatformWeb               = "web"
	PlatformMobile            = "mobile"
)

func AuthenticateMiddleware() func(*gin.Context) {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader(CtxAuthorization))
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			response.JSON(c, http.StatusUnauthorized, false, "Unauthorized", nil)
			return
		}

		platform := strings.ToLower(strings.TrimSpace(c.GetHeader(CtxPlatform)))
		if platform != PlatformWeb && platform != PlatformMobile {
			response.JSON(c, http.StatusBadRequest, false, "Invalid platform", nil)
			return
		}

		accessToken := strings.TrimSpace(authHeader[7:])

		userID, name, tokenPlatform, err := utils.ParseJWT(accessToken)
		if err != nil {
			response.JSON(c, http.StatusBadRequest, false, "Unauthorized", nil)
			return
		}

		if tokenPlatform != platform {
			response.JSON(c, http.StatusUnauthorized, false, "Unauthorized", nil)
			return
		}

		c.Set(CtxUserID, userID)
		c.Set(CtxUserDisplayName, name)
		c.Set(CtxPlatform, platform)
	}
}
