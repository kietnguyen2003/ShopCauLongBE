package notification

import (
	"fmt"
	domainNotification "kafka-order-demo/backend/internal/domain/notification"
)

type Service struct {
	notificationRepo NotificationRepository
	broadcaster      Broadcaster
}

func NewService(notificationRepo NotificationRepository, broadcaster Broadcaster) *Service {
	return &Service{
		notificationRepo: notificationRepo,
		broadcaster:      broadcaster,
	}
}

func (s *Service) CreateOrderStatusNotification(userID, orderID uint, status string) (*NotificationResponse, error) {
	notificationType, title, message, ok := orderStatusNotificationContent(orderID, status)
	if !ok {
		return nil, nil
	}

	data := fmt.Sprintf(`{"order_id":%d,"status":"%s"}`, orderID, status)
	notification, err := domainNotification.NewNotification(userID, notificationType, title, message, data)
	if err != nil {
		return nil, err
	}
	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, err
	}

	response := toNotificationResponse(notification)
	if s.broadcaster != nil {
		s.broadcaster.BroadcastToUser(userID, response)
	}

	return &response, nil
}

func (s *Service) GetNotifications(userID uint, page, limit int) (*NotificationListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	notifications, total, err := s.notificationRepo.GetByUserID(userID, page, limit)
	if err != nil {
		return nil, err
	}

	return &NotificationListResponse{
		Items:      toNotificationResponses(notifications),
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages(total, limit),
	}, nil
}

func (s *Service) CountUnread(userID uint) (*UnreadCountResponse, error) {
	count, err := s.notificationRepo.CountUnreadByUserID(userID)
	if err != nil {
		return nil, err
	}

	return &UnreadCountResponse{Count: count}, nil
}

func (s *Service) MarkAsRead(id, userID uint) (*NotificationResponse, error) {
	notification, err := s.notificationRepo.GetByIDAndUserID(id, userID)
	if err != nil {
		return nil, err
	}

	notification.MarkAsRead()
	if err := s.notificationRepo.Update(notification); err != nil {
		return nil, err
	}

	response := toNotificationResponse(notification)
	return &response, nil
}

func (s *Service) MarkAllAsRead(userID uint) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

func orderStatusNotificationContent(orderID uint, status string) (string, string, string, bool) {
	switch status {
	case "confirmed":
		return domainNotification.TypeOrderConfirmed,
			"Order confirmed",
			fmt.Sprintf("Your order #%d has been confirmed.", orderID),
			true
	case "cancelled":
		return domainNotification.TypeOrderCancelled,
			"Order cancelled",
			fmt.Sprintf("Your order #%d has been cancelled.", orderID),
			true
	default:
		return "", "", "", false
	}
}
