package notification

import (
	"errors"
	"strings"
	"time"
)

const (
	TypeOrderConfirmed = "order_confirmed"
	TypeOrderCancelled = "order_cancelled"
)

var (
	ErrUserIDRequired      = errors.New("user ID cannot be zero")
	ErrNotificationType    = errors.New("notification type is required")
	ErrNotificationTitle   = errors.New("notification title is required")
	ErrNotificationMessage = errors.New("notification message is required")
)

type Notification struct {
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

func NewNotification(userID uint, notificationType, title, message, data string) (*Notification, error) {
	now := time.Now()
	notification := &Notification{
		UserID:    userID,
		Type:      strings.TrimSpace(notificationType),
		Title:     strings.TrimSpace(title),
		Message:   strings.TrimSpace(message),
		Data:      strings.TrimSpace(data),
		IsRead:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := notification.Validate(); err != nil {
		return nil, err
	}

	return notification, nil
}

func (n *Notification) Validate() error {
	if n.UserID == 0 {
		return ErrUserIDRequired
	}
	if n.Type == "" {
		return ErrNotificationType
	}
	if n.Title == "" {
		return ErrNotificationTitle
	}
	if n.Message == "" {
		return ErrNotificationMessage
	}

	return nil
}

func (n *Notification) MarkAsRead() {
	if n.IsRead {
		return
	}
	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
	n.UpdatedAt = now
}
