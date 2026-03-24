package realtime

import (
	"log"
	"sync"
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

func (h *Hub) SendEventToUserIDs(userIDs []int64, senderID int64, eventType EventType, payload string) {
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
