package http

import (
	appNotification "kafka-order-demo/backend/internal/application/notification"
	"sync"

	"github.com/gorilla/websocket"
)

type NotificationHub struct {
	mu      sync.RWMutex
	clients map[uint]map[*websocket.Conn]struct{}
}

func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients: make(map[uint]map[*websocket.Conn]struct{}),
	}
}

func (h *NotificationHub) AddClient(userID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[*websocket.Conn]struct{})
	}
	h.clients[userID][conn] = struct{}{}
}

func (h *NotificationHub) RemoveClient(userID uint, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		return
	}
	delete(h.clients[userID], conn)
	if len(h.clients[userID]) == 0 {
		delete(h.clients, userID)
	}
}

func (h *NotificationHub) BroadcastToUser(userID uint, notification appNotification.NotificationResponse) {
	h.mu.RLock()
	connections := make([]*websocket.Conn, 0, len(h.clients[userID]))
	for conn := range h.clients[userID] {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	payload := toNotificationHTTPResponse(notification)
	for _, conn := range connections {
		if err := conn.WriteJSON(payload); err != nil {
			h.RemoveClient(userID, conn)
			_ = conn.Close()
		}
	}
}
