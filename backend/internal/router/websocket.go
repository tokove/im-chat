package router

import (
	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/realtime"
	"backend/pkg/response"
	"backend/pkg/utils"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/gin-gonic/gin"
)

func handleWebSocket(hub *realtime.Hub, c *gin.Context) {
	authHeader := c.GetHeader(middleware.CtxAuthorization)
	if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		response.JSON(c, http.StatusUnauthorized, false, "unauthorization", nil)
		return
	}

	accessToken := strings.TrimSpace(authHeader[7:])
	if accessToken == "" {
		response.JSON(c, http.StatusUnauthorized, false, "unauthorization", nil)
		return
	}

	userID, _, _, err := utils.ParseJWT(accessToken)
	if err != nil {
		response.JSON(c, http.StatusUnauthorized, false, "invalid token", nil)
		return
	}

	user, err := model.GetUserByID(userID)
	if err != nil {
		response.JSON(c, http.StatusUnauthorized, false, "user not found", nil)
		return
	}

	w, r := c.Writer, c.Request
	opts := websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	}
	conn, err := websocket.Accept(w, r, &opts)
	if err != nil {
		response.JSON(c, http.StatusInternalServerError, false, "failed to upgrade websocket", nil)
		return
	}

	client := realtime.NewClient(user, conn)

	hub.RegisterClientConnection(client)
	hub.SendCurrentClients(client)

	defer func() {
		hub.UnregisterClientConnection(client)
		client.Close()
	}()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go heartbeat(ctx, cancel, client)
	go writePump(ctx, client)
	readPump(ctx, cancel, hub, client)
}

func heartbeat(ctx context.Context, cancel context.CancelFunc, client *realtime.Client) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
			if err := client.Conn.Ping(pingCtx); err != nil {
				log.Println("ping failed, disconnecting client")
				pingCancel()
				cancel()
				return
			}
			pingCancel()

			client.Send <- realtime.Event{
				EventType: realtime.EventHeartbeat,
				Payload:   nil,
			}
		}
	}
}

func writePump(ctx context.Context, client *realtime.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-client.Send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

			_ = wsjson.Write(writeCtx, client.Conn, event)
			cancel()
		}
	}
}

func readPump(ctx context.Context, cancel context.CancelFunc, hub *realtime.Hub, client *realtime.Client) {
	defer cancel()
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Recovered from panic in readPump for client %d: %v", client.User.ID, r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		var event realtime.Event
		if err := wsjson.Read(ctx, client.Conn, &event); err != nil {
			return
		}

		handleIncomingEvent(hub, client, event)
	}
}

func handleEventMessage(payload map[string]any, hub *realtime.Hub, client *realtime.Client) {
	privateIdAny, ok := payload["private_id"]
	if !ok {
		hub.SendError(client.User.ID, "private id is missing")
		return
	}
	privateIdFloat, ok := privateIdAny.(float64)
	if !ok {
		hub.SendError(client.User.ID, "private id must be a number")
		return
	}
	privateId := int64(privateIdFloat)

	receiverIdAny, ok := payload["receiver_id"]
	if !ok {
		hub.SendError(client.User.ID, "receiver id is missing")
		return
	}
	receiverIdFloat, ok := receiverIdAny.(float64)
	if !ok {
		hub.SendError(client.User.ID, "receiver id must be a number")
		return
	}
	receiverId := int64(receiverIdFloat)

	messageTypeAny, ok := payload["message_type"]
	if !ok {
		hub.SendError(client.User.ID, "message_type is missing")
		return
	}
	messageType, ok := messageTypeAny.(string)
	if !ok {
		hub.SendError(client.User.ID, "message_type must be a string")
		return
	}

	msgBytes, err := json.Marshal(payload)
	if err != nil {
		hub.SendError(client.User.ID, "invalid message payload")
		return
	}
	var msg model.Message
	if err := json.Unmarshal(msgBytes, &msg); err != nil {
		hub.SendError(client.User.ID, "invalid message format")
		return
	}

	msg.ID = 0
	msg.FromID = client.User.ID
	msg.PrivateID = privateId
	msg.MessageType = messageType
	msg.CreatedAt = time.Now()

	if err = model.CreateMessage(&msg); err != nil {
		hub.SendError(client.User.ID, "failed to save message")
		return
	}

	hub.SendEventToUserIDs([]int64{msg.FromID, receiverId}, msg.FromID, realtime.EventMessage, map[string]any{
		"message": msg,
	})
}

func handleIncomingEvent(hub *realtime.Hub, client *realtime.Client, event realtime.Event) {
	payload, ok := event.Payload.(map[string]any)
	if !ok {
		hub.SendError(client.User.ID, "payload not exists")
		return
	}

	switch event.EventType {
	case realtime.EventMessage:
		handleEventMessage(payload, hub, client)
	case realtime.EventDelivered:
		msgIDAny, ok := payload["message_id"]
		if !ok {
			hub.SendError(client.User.ID, "message id not exists")
			return
		}
		msgIDFloat, ok := msgIDAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "message id type is wrong")
			return
		}
		msgID := int64(msgIDFloat)
		msg, err := model.GetMessageByID(msgID)
		if err != nil {
			hub.SendError(client.User.ID, "message not exists")
			return
		}

		if msg.FromID == client.User.ID {
			hub.SendError(client.User.ID, "can't send message for self")
			return
		}

		if err := model.MarkMessageDelivered(msgID); err != nil {
			hub.SendError(client.User.ID, "message deliver failed")
			return
		}
		hub.SendEventToUserIDs([]int64{msg.FromID}, client.User.ID, realtime.EventDelivered, map[string]any{
			"message_id": msgID,
			"to_id":      client.User.ID,
		})

	case realtime.EventRead:
		msgIDAny, ok := payload["message_id"]
		if !ok {
			hub.SendError(client.User.ID, "message id not exists")
			return
		}
		msgIDFloat, ok := msgIDAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "message id type is wrong")
			return
		}
		msgID := int64(msgIDFloat)
		msg, err := model.GetMessageByID(msgID)
		if err != nil {
			hub.SendError(client.User.ID, "message not exists")
			return
		}

		if msg.FromID == client.User.ID {
			hub.SendError(client.User.ID, "can't send message for self")
			return
		}

		if err := model.MarkMessageRead(msgID); err != nil {
			hub.SendError(client.User.ID, "message read failed")
			return
		}
		hub.SendEventToUserIDs([]int64{msg.FromID}, client.User.ID, realtime.EventRead, map[string]any{
			"message_id": msgID,
		})
		
	case realtime.EventTyping:
		privateIDAny, ok := payload["private_id"]
		if !ok {
			hub.SendError(client.User.ID, "private id not exists")
			return
		}
		privateIDFloat, ok := privateIDAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "private id type is wrong")
			return
		}
		privateID := int64(privateIDFloat)

		reciverIDAny, ok := payload["receiver_id"]
		if !ok {
			hub.SendError(client.User.ID, "receiver id not exists")
			return
		}
		receiverIDFloat, ok := reciverIDAny.(float64)
		if !ok {
			hub.SendError(client.User.ID, "receiver id type is wrong")
			return
		}
		receiverID := int64(receiverIDFloat)

		isTypingAny, ok := payload["is_typing"]
		if !ok {
			hub.SendError(client.User.ID, "isTyping not exists")
			return
		}
		isTyping, ok := isTypingAny.(bool)
		if !ok {
			hub.SendError(client.User.ID, "isTyping type is wrong")
			return
		}

		hub.SendEventToUserIDs([]int64{receiverID}, client.User.ID, realtime.EventTyping, map[string]any{
			"private_id": privateID,
			"user_id":    client.User.ID,
			"is_typing":  isTyping,
		})

	default:
		hub.SendError(client.User.ID, "unknown event type: "+string(event.EventType))
	}
}
