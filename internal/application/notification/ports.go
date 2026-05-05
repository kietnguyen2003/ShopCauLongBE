package notification

import domainNotification "kafka-order-demo/backend/internal/domain/notification"

type NotificationRepository interface {
	Create(notification *domainNotification.Notification) error
	GetByUserID(userID uint, page, limit int) ([]*domainNotification.Notification, int64, error)
	GetByIDAndUserID(id, userID uint) (*domainNotification.Notification, error)
	CountUnreadByUserID(userID uint) (int64, error)
	Update(notification *domainNotification.Notification) error
	MarkAllAsRead(userID uint) error
}

type Broadcaster interface {
	BroadcastToUser(userID uint, notification NotificationResponse)
}
