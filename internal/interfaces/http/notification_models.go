package http

import (
	appNotification "kafka-order-demo/backend/internal/application/notification"
	"time"
)

type notificationResponse struct {
	ID        uint       `json:"id"`
	UserID    uint       `json:"user_id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Message   string     `json:"message"`
	Data      string     `json:"data"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type notificationListResponse struct {
	Items      []notificationResponse `json:"items"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	Total      int64                  `json:"total"`
	TotalPages int64                  `json:"total_pages"`
}

type unreadCountResponse struct {
	Count int64 `json:"count"`
}

func toNotificationHTTPResponse(resp appNotification.NotificationResponse) notificationResponse {
	return notificationResponse{
		ID:        resp.ID,
		UserID:    resp.UserID,
		Type:      resp.Type,
		Title:     resp.Title,
		Message:   resp.Message,
		Data:      resp.Data,
		IsRead:    resp.IsRead,
		ReadAt:    resp.ReadAt,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func toNotificationListHTTPResponse(resp *appNotification.NotificationListResponse) notificationListResponse {
	items := make([]notificationResponse, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = toNotificationHTTPResponse(item)
	}

	return notificationListResponse{
		Items:      items,
		Page:       resp.Page,
		Limit:      resp.Limit,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
	}
}

func toUnreadCountHTTPResponse(resp *appNotification.UnreadCountResponse) unreadCountResponse {
	return unreadCountResponse{Count: resp.Count}
}
