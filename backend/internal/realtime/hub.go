package realtime

import (
	"backend/internal/model"
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
)

type Hub struct {
	Clients map[int64]map[*Client]struct{}
	mtx     sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients: make(map[int64]map[*Client]struct{}),
	}
}

func (h *Hub) broadcastToAll(event Event) {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	for _, conns := range h.Clients {
		for c := range conns {
			select {
			case c.Send <- event:
			default:
				log.Printf("drop event for client %d, chan full", c.User.ID)
			}
		}
	}
}

func (h *Hub) GetClients(userID int64) ([]*Client, bool) {
	h.mtx.RLock()
	defer h.mtx.RUnlock()

	conns, ok := h.Clients[userID]
	if !ok || len(conns) == 0 {
		return nil, false
	}

	clients := make([]*Client, 0, len(conns))
	for c := range conns {
		clients = append(clients, c)
	}

	return clients, true
}

func (h *Hub) SendEventToUserIDs(userIDs []int64, senderID int64, eventType EventType, payload any) {
	for _, id := range userIDs {
		h.mtx.RLock()
		conns, ok := h.Clients[id]
		h.mtx.RUnlock()

		if !ok {
			continue
		}

		for c := range conns {
			c.SendEvent(Event{
				EventType: eventType,
				Payload:   payload,
			})
		}
	}
}

func (h *Hub) RegisterClientConnection(client *Client) {
	h.mtx.Lock()
	conns, ok := h.Clients[client.User.ID]
	if !ok {
		conns = make(map[*Client]struct{})
		h.Clients[client.User.ID] = conns
	}
	conns[client] = struct{}{}
	firstConnection := len(conns) == 1
	h.mtx.Unlock()

	if firstConnection {
		h.broadcastToAll(Event{
			EventType: EventUserOnline,
			Payload:   client.User.ToMap(),
		})

		go func() {
			privates, err := model.GetPrivatesForUser(client.User.ID)
			if err != nil {
				fmt.Printf("RegisterClientConnection: GetPrivatesForUser, err: %v", err)
				return
			}

			for _, p := range privates {
				msgs, err := model.GetUndeliveredMessagesByPrivateID(p.ID)
				if err != nil {
					fmt.Printf("RegisterClientConnection: GetUndeliveredMessagesByPrivateID, err: %v", err)
					continue
				}

				for _, m := range msgs {
					if m.FromID == client.User.ID {
						continue
					}
					h.SendEventToUserIDs([]int64{m.FromID}, client.User.ID, EventDelivered, gin.H{
						"message_id": m.ID,
						"to_id":      client.User.ID,
					})
				}
			}
		}()
	}
}

func (h *Hub) UnregisterClientConnection(client *Client) {
	h.mtx.Lock()
	conns, ok := h.Clients[client.User.ID]
	if !ok {
		h.mtx.Unlock()
		return
	}

	delete(conns, client)

	noConnectionLeft := len(conns) == 0
	if noConnectionLeft {
		delete(h.Clients, client.User.ID)
	}
	h.mtx.Unlock()

	if err := client.Conn.Close(); err != nil {
		fmt.Printf("UnregisterClientConnection: client.Conn.Close(), err: %v", err)
		return
	}

	if noConnectionLeft {
		h.broadcastToAll(Event{
			EventType: EventUserOffline,
			Payload:   client.User.ToMap(),
		})
	}
}

func (h *Hub) SendCurrentClients(toClient *Client) {
	h.mtx.RLock()

	users := []map[string]any{}
	seen := make(map[int64]struct{})

	for userID, conns := range h.Clients {
		if userID == toClient.User.ID {
			continue
		}

		if _, ok := seen[userID]; ok {
			continue
		}

		for c := range conns {
			users = append(users, c.User.ToMap())
			seen[c.User.ID] = struct{}{}
			break
		}
	}

	h.mtx.RUnlock()

	toClient.Send <- Event{
		EventType: EventCurrentUsers,
		Payload:   users,
	}
}

func (h *Hub) SendError(clientID int64, message string) {
	clients, ok := h.GetClients(clientID)
	if !ok || len(clients) == 0 {
		log.Println("SendError: GetClients, client not exists")
		return
	}

	for _, c := range clients {
		c.SendEvent(Event{
			EventType: EventError,
			Payload:   message,
		})
	}
}

func (h *Hub) Shutdown() {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	log.Println("Shutting down hub, notify all clients ...")
	for _, conns := range h.Clients {
		for c := range conns {
			c.SendEvent(Event{
				EventType: EventServerShutdown,
				Payload:   "Server is shutting down",
			})
			c.Conn.Close()
		}
	}

	h.Clients = make(map[int64]map[*Client]struct{})
	log.Println("Hub shutdown complete")
}
