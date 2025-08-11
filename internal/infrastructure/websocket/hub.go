package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo
	},
}

type Hub struct {
	// Registered user clients
	userClients map[uint]*websocket.Conn

	// Registered admin clients
	adminClients map[*websocket.Conn]bool

	// Register requests from user clients
	registerUser chan UserConnection

	// Register requests from admin clients
	registerAdmin chan *websocket.Conn

	// Unregister requests from user clients
	unregisterUser chan UserConnection

	// Unregister requests from admin clients
	unregisterAdmin chan *websocket.Conn

	// Message broadcast to specific user
	userMessages chan UserMessage

	// Message broadcast to all admins
	adminMessages chan Message
}

type UserConnection struct {
	UserID uint
	Conn   *websocket.Conn
}

type UserMessage struct {
	UserID  uint
	Message Message
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewHub() *Hub {
	return &Hub{
		userClients:     make(map[uint]*websocket.Conn),
		adminClients:    make(map[*websocket.Conn]bool),
		registerUser:    make(chan UserConnection),
		registerAdmin:   make(chan *websocket.Conn),
		unregisterUser:  make(chan UserConnection),
		unregisterAdmin: make(chan *websocket.Conn),
		userMessages:    make(chan UserMessage),
		adminMessages:   make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case userConn := <-h.registerUser:
			h.userClients[userConn.UserID] = userConn.Conn
			log.Printf("User %d connected", userConn.UserID)

		case adminConn := <-h.registerAdmin:
			h.adminClients[adminConn] = true
			log.Println("Admin connected")

		case userConn := <-h.unregisterUser:
			if conn, ok := h.userClients[userConn.UserID]; ok {
				delete(h.userClients, userConn.UserID)
				conn.Close()
				log.Printf("User %d disconnected", userConn.UserID)
			}

		case adminConn := <-h.unregisterAdmin:
			if _, ok := h.adminClients[adminConn]; ok {
				delete(h.adminClients, adminConn)
				adminConn.Close()
				log.Println("Admin disconnected")
			}

		case userMessage := <-h.userMessages:
			if conn, ok := h.userClients[userMessage.UserID]; ok {
				if err := conn.WriteJSON(userMessage.Message); err != nil {
					log.Printf("Error sending message to user %d: %v", userMessage.UserID, err)
					conn.Close()
					delete(h.userClients, userMessage.UserID)
				}
			}

		case adminMessage := <-h.adminMessages:
			for conn := range h.adminClients {
				if err := conn.WriteJSON(adminMessage); err != nil {
					log.Printf("Error sending message to admin: %v", err)
					conn.Close()
					delete(h.adminClients, conn)
				}
			}
		}
	}
}

func (h *Hub) SendToUser(userID uint, message Message) {
	select {
	case h.userMessages <- UserMessage{UserID: userID, Message: message}:
	default:
		log.Printf("Failed to send message to user %d: channel full", userID)
	}
}

func (h *Hub) BroadcastToAdmin(message Message) {
	select {
	case h.adminMessages <- message:
	default:
		log.Println("Failed to broadcast to admin: channel full")
	}
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) HandleUserConnection(c *gin.Context) {
	userIDParam := c.Param("userID")
	userID, err := strconv.ParseUint(userIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	userConn := UserConnection{
		UserID: uint(userID),
		Conn:   conn,
	}

	h.hub.registerUser <- userConn

	// Handle disconnection
	defer func() {
		h.hub.unregisterUser <- userConn
	}()

	// Keep connection alive and handle incoming messages
	for {
		var msg json.RawMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for user %d: %v", userID, err)
			}
			break
		}
		// Echo back or handle user messages if needed
		log.Printf("Received message from user %d: %s", userID, string(msg))
	}
}

func (h *Handler) HandleAdminConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade admin connection: %v", err)
		return
	}

	h.hub.registerAdmin <- conn

	// Handle disconnection
	defer func() {
		h.hub.unregisterAdmin <- conn
	}()

	// Keep connection alive and handle incoming messages
	for {
		var msg json.RawMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for admin: %v", err)
			}
			break
		}
		// Echo back or handle admin messages if needed
		log.Printf("Received message from admin: %s", string(msg))
	}
}