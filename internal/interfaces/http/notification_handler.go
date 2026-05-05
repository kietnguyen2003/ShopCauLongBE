package http

import (
	"net/http"
	"strconv"

	appNotification "kafka-order-demo/backend/internal/application/notification"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type NotificationHandler struct {
	notificationService *appNotification.Service
	hub                 *NotificationHub
	upgrader            websocket.Upgrader
}

func NewNotificationHandler(notificationService *appNotification.Service, hub *NotificationHub) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		hub:                 hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	page, err := parsePositiveIntQuery(c, "page", 1)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	limit, err := parsePositiveIntQuery(c, "limit", 20)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.notificationService.GetNotifications(userID.(uint), page, limit)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get notifications successfully", toNotificationListHTTPResponse(resp))
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	resp, err := h.notificationService.CountUnread(userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get unread notification count successfully", toUnreadCountHTTPResponse(resp))
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	resp, err := h.notificationService.MarkAsRead(uint(id), userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Mark notification as read successfully", toNotificationHTTPResponse(*resp))
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	if err := h.notificationService.MarkAllAsRead(userID.(uint)); err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Mark all notifications as read successfully", nil)
}

func (h *NotificationHandler) StreamNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.AddClient(userID.(uint), conn)
	defer func() {
		h.hub.RemoveClient(userID.(uint), conn)
		_ = conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}
