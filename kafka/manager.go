package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Manager struct {
	brokers []string
	writers map[string]*kafka.Writer
	readers map[string]*kafka.Reader
}

type OrderEvent struct {
	OrderID     uint      `json:"order_id"`
	UserID      uint      `json:"user_id"`
	Type        string    `json:"type"` // "created", "updated"
	Status      string    `json:"status"`
	TotalAmount float64   `json:"total_amount"`
	Timestamp   time.Time `json:"timestamp"`
}

type NotificationEvent struct {
	UserID    uint      `json:"user_id,omitempty"`
	Type      string    `json:"type"` // "order_created", "order_updated", "admin_notification"
	Message   string    `json:"message"`
	OrderID   uint      `json:"order_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

const (
	TopicOrderCreated        = "order-created"
	TopicOrderUpdated        = "order-updated"
	TopicUserNotifications   = "user-notifications"
	TopicAdminNotifications  = "admin-notifications"
)

func NewManager(brokers []string) *Manager {
	m := &Manager{
		brokers: brokers,
		writers: make(map[string]*kafka.Writer),
		readers: make(map[string]*kafka.Reader),
	}

	// Initialize writers
	topics := []string{TopicOrderCreated, TopicOrderUpdated, TopicUserNotifications, TopicAdminNotifications}
	for _, topic := range topics {
		m.writers[topic] = &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
	}

	return m
}

func (m *Manager) PublishOrderEvent(event OrderEvent) error {
	writer, ok := m.writers[getTopicForOrderEvent(event.Type)]
	if !ok {
		log.Printf("No writer found for event type: %s", event.Type)
		return nil
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: eventJSON,
		},
	)
}

func (m *Manager) PublishNotification(event NotificationEvent) error {
	var topic string
	switch event.Type {
	case "admin_notification":
		topic = TopicAdminNotifications
	default:
		topic = TopicUserNotifications
	}

	writer, ok := m.writers[topic]
	if !ok {
		log.Printf("No writer found for notification type: %s", event.Type)
		return nil
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: eventJSON,
		},
	)
}

func (m *Manager) StartConsumers() {
	// Start consumer for order events
	go m.consumeOrderEvents()
	
	// Start consumer for user notifications
	go m.consumeUserNotifications()
	
	// Start consumer for admin notifications
	go m.consumeAdminNotifications()
}

func (m *Manager) consumeOrderEvents() {
	// Consumer for order created events
	createdReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  m.brokers,
		Topic:    TopicOrderCreated,
		GroupID:  "order-processor",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer createdReader.Close()

	// Consumer for order updated events
	updatedReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  m.brokers,
		Topic:    TopicOrderUpdated,
		GroupID:  "order-processor",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer updatedReader.Close()

	// Process created orders
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			msg, err := createdReader.ReadMessage(ctx)
			cancel()
			if err != nil {
				if err != context.DeadlineExceeded {
					log.Printf("Error reading order created message: %v", err)
				}
				time.Sleep(1 * time.Second)
				continue
			}

			var event OrderEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling order created event: %v", err)
				continue
			}

			log.Printf("Processing order created event: OrderID=%d, UserID=%d", event.OrderID, event.UserID)
			
			// Send notification to admin
			adminNotification := NotificationEvent{
				Type:      "admin_notification",
				Message:   "Đơn hàng mới được tạo",
				OrderID:   event.OrderID,
				Timestamp: time.Now(),
			}
			m.PublishNotification(adminNotification)
		}
	}()

	// Process updated orders
	go func() {
		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			msg, err := updatedReader.ReadMessage(ctx)
			cancel()
			if err != nil {
				if err != context.DeadlineExceeded {
					log.Printf("Error reading order updated message: %v", err)
				}
				time.Sleep(1 * time.Second)
				continue
			}

			var event OrderEvent
			if err := json.Unmarshal(msg.Value, &event); err != nil {
				log.Printf("Error unmarshaling order updated event: %v", err)
				continue
			}

			log.Printf("Processing order updated event: OrderID=%d, Status=%s", event.OrderID, event.Status)
			
			// Send notification to user
			var message string
			switch event.Status {
			case "confirmed":
				message = "Đơn hàng của bạn đã được xác nhận"
			case "shipped":
				message = "Đơn hàng của bạn đã được giao vận"
			case "delivered":
				message = "Đơn hàng của bạn đã được giao thành công"
			case "cancelled":
				message = "Đơn hàng của bạn đã bị hủy"
			default:
				message = "Trạng thái đơn hàng của bạn đã thay đổi"
			}

			userNotification := NotificationEvent{
				UserID:    event.UserID,
				Type:      "order_updated",
				Message:   message,
				OrderID:   event.OrderID,
				Timestamp: time.Now(),
			}
			m.PublishNotification(userNotification)
		}
	}()
}

func (m *Manager) consumeUserNotifications() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  m.brokers,
		Topic:    TopicUserNotifications,
		GroupID:  "user-notification-processor",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msg, err := reader.ReadMessage(ctx)
		cancel()
		if err != nil {
			if err != context.DeadlineExceeded {
				log.Printf("Error reading user notification message: %v", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		var event NotificationEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshaling user notification event: %v", err)
			continue
		}

		log.Printf("Processing user notification: UserID=%d, Message=%s", event.UserID, event.Message)
		// Here you would typically send to WebSocket or store in database
	}
}

func (m *Manager) consumeAdminNotifications() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  m.brokers,
		Topic:    TopicAdminNotifications,
		GroupID:  "admin-notification-processor",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		msg, err := reader.ReadMessage(ctx)
		cancel()
		if err != nil {
			if err != context.DeadlineExceeded {
				log.Printf("Error reading admin notification message: %v", err)
			}
			time.Sleep(1 * time.Second)
			continue
		}

		var event NotificationEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("Error unmarshaling admin notification event: %v", err)
			continue
		}

		log.Printf("Processing admin notification: Message=%s", event.Message)
		// Here you would typically send to WebSocket or store in database
	}
}

func (m *Manager) Close() {
	for _, writer := range m.writers {
		writer.Close()
	}
	for _, reader := range m.readers {
		reader.Close()
	}
}

func getTopicForOrderEvent(eventType string) string {
	switch eventType {
	case "created":
		return TopicOrderCreated
	case "updated":
		return TopicOrderUpdated
	default:
		return TopicOrderCreated
	}
}