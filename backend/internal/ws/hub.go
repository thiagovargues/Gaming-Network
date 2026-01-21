package ws

import (
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int64]map[*Client]struct{})}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.UserID] == nil {
		h.clients[client.UserID] = make(map[*Client]struct{})
	}
	h.clients[client.UserID][client] = struct{}{}
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.clients[client.UserID]
	if clients == nil {
		return
	}
	delete(clients, client)
	if len(clients) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *Hub) SendToUser(userID int64, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients := h.clients[userID]
	for c := range clients {
		select {
		case c.Send <- payload:
		default:
		}
	}
}
