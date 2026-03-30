package router

import (
	"backend/pkg/response"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
)

func handleHealthCheckHTTP(c *gin.Context) {
	response.JSON(c, http.StatusOK, true, "API is healthy", nil)
}

func handleHealthCheckWs(c *gin.Context) {
	// build conn
	w, r := c.Writer, c.Request
	opts := &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "Connection closed")

	ctx := r.Context()
	for {
		var message string
		if err := wsjson.Read(ctx, conn, &message); err != nil {
			break
		}

		response := map[string]any{
			"data":    message,
			"from":    "server",
			"success": true,
		}

		if err = wsjson.Write(ctx, conn, response); err != nil {
			break
		}
	}
}
