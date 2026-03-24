package realtime

import (
	"backend/internal/model"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	User *model.User     `json:"user"`
	Conn *websocket.Conn `json:"-"`
	Send chan Event      `json:"-"`
	once sync.Once       `json:"-"`
}

func NewClient(user *model.User, conn *websocket.Conn) *Client {
	return &Client{
		User: user,
		Conn: conn,
		Send: make(chan Event, 512),
	}
}

func (c *Client) SendEvent(event Event) {
	select {
	case c.Send <- event:
	default:
	}
}

func (c *Client) Close() {
	c.once.Do(func() {
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
		close(c.Send)
	})
}
