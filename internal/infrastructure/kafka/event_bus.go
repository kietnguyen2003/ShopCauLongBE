package kafka

import (
	"encoding/json"
	"log"
	
	"kafka-order-demo/backend/internal/domain/order"
	"kafka-order-demo/backend/internal/infrastructure/websocket"
	"github.com/IBM/sarama"
)

type EventBus struct {
	producer sarama.SyncProducer
	hub      *websocket.Hub
}

func NewEventBus(brokers []string, hub *websocket.Hub) (*EventBus, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &EventBus{
		producer: producer,
		hub:      hub,
	}, nil
}

func (eb *EventBus) PublishOrderCreated(ord *order.Order) error {
	// Create message for Kafka
	orderData := map[string]interface{}{
		"order_id":      ord.ID,
		"user_id":       ord.UserID,
		"customer_name": ord.CustomerName,
		"phone":         ord.Phone,
		"address":       ord.Address,
		"email":         ord.Email,
		"total_amount":  ord.TotalAmount,
		"status":        ord.Status,
		"items":         ord.OrderItems,
		"created_at":    ord.CreatedAt,
	}

	jsonData, err := json.Marshal(orderData)
	if err != nil {
		log.Printf("Error marshaling order created event: %v", err)
		return err
	}

	// Send to Kafka
	msg := &sarama.ProducerMessage{
		Topic: "order-created",
		Value: sarama.StringEncoder(jsonData),
	}

	_, _, err = eb.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Error sending order created event to Kafka: %v", err)
	}

	// Send WebSocket notification to admin
	if eb.hub != nil {
		wsMessage := websocket.Message{
			Type: "new_order",
			Data: orderData,
		}
		eb.hub.BroadcastToAdmin(wsMessage)
	}

	return err
}

func (eb *EventBus) PublishOrderUpdated(ord *order.Order) error {
	// Create message for Kafka
	orderData := map[string]interface{}{
		"order_id":     ord.ID,
		"user_id":      ord.UserID,
		"status":       ord.Status,
		"total_amount": ord.TotalAmount,
		"updated_at":   ord.UpdatedAt,
	}

	jsonData, err := json.Marshal(orderData)
	if err != nil {
		log.Printf("Error marshaling order updated event: %v", err)
		return err
	}

	// Send to Kafka
	msg := &sarama.ProducerMessage{
		Topic: "order-updated",
		Value: sarama.StringEncoder(jsonData),
	}

	_, _, err = eb.producer.SendMessage(msg)
	if err != nil {
		log.Printf("Error sending order updated event to Kafka: %v", err)
	}

	// Send WebSocket notification to user
	if eb.hub != nil {
		wsMessage := websocket.Message{
			Type: "order_status_updated",
			Data: orderData,
		}
		eb.hub.SendToUser(ord.UserID, wsMessage)
	}

	return err
}

func (eb *EventBus) Close() error {
	return eb.producer.Close()
}