package notification

import (
	domainNotification "kafka-order-demo/backend/internal/domain/notification"
	"time"
)

type NotificationResponse struct {
	ID        uint
	UserID    uint
	Type      string
	Title     string
	Message   string
	Data      string
	IsRead    bool
	ReadAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NotificationListResponse struct {
	Items      []NotificationResponse
	Page       int
	Limit      int
	Total      int64
	TotalPages int64
}

type UnreadCountResponse struct {
	Count int64
}

func toNotificationResponse(notification *domainNotification.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        notification.ID,
		UserID:    notification.UserID,
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Data:      notification.Data,
		IsRead:    notification.IsRead,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
		UpdatedAt: notification.UpdatedAt,
	}
}

func toNotificationResponses(notifications []*domainNotification.Notification) []NotificationResponse {
	responses := make([]NotificationResponse, len(notifications))
	for i, notification := range notifications {
		responses[i] = toNotificationResponse(notification)
	}
	return responses
}

func totalPages(total int64, limit int) int64 {
	if limit <= 0 {
		return 0
	}
	pages := total / int64(limit)
	if total%int64(limit) != 0 {
		pages++
	}
	return pages
}
