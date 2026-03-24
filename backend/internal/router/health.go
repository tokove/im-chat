package router

import (
	"backend/pkg/response"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func handleHealthCheckHTTP(c *gin.Context) {
	response.JSON(c, http.StatusOK, true, "API is healthy", nil)
}

func handleHealthCheckWs(c *gin.Context) {
	// http -> websocket
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// build conn
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		// read message
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("conn.ReadMessage failed, err: %v", err)
			break
		}

		response := map[string]any{
			"data":    string(message),
			"from":    "server",
			"success": true,
		}

		log.Println("WS message received:", string(message))

		// write message
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("conn.WriteJSON failed, err: %v", err)
			break
		}
	}
}
