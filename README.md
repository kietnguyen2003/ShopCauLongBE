# Backend - Apache Kafka Implementation

Chi tiết implementation của Apache Kafka trong Go backend cho hệ thống đặt hàng real-time.

## 🏗️ Kiến trúc Kafka

```
Order API ──> Kafka Producer ──> Topics ──> Kafka Consumer ──> WebSocket
    │                                │                              │
    │         ┌─ order-created       │         ┌─ Process Events   │
    │         ├─ order-updated       │         ├─ Send Notifications
    │         ├─ user-notifications  │         └─ Broadcast WebSocket
    │         └─ admin-notifications │
    │                                │
    └─ Direct WebSocket Broadcast ───┘
```

## 📡 Kafka Topics

### 1. `order-created` 
**Mục đích**: Thông báo khi có đơn hàng mới được tạo

**Producer**: `handlers/orders.go:CreateOrder()`
```go
orderEvent := kafka.OrderEvent{
    OrderID:     order.ID,
    UserID:      userID,
    Type:        "created",
    Status:      order.Status,
    TotalAmount: order.TotalAmount,
    Timestamp:   time.Now(),
}
kafkaManager.PublishOrderEvent(orderEvent)
```

**Consumer**: `kafka/manager.go:consumeOrderEvents()`
```go
// Process created orders
go func() {
    for {
        msg, err := createdReader.ReadMessage(ctx)
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
```

### 2. `order-updated`
**Mục đích**: Thông báo khi trạng thái đơn hàng thay đổi

**Producer**: `handlers/orders.go:UpdateOrderStatus()`
```go
orderEvent := kafka.OrderEvent{
    OrderID:     order.ID,
    UserID:      order.UserID,
    Type:        "updated",
    Status:      req.Status,
    TotalAmount: order.TotalAmount,
    Timestamp:   time.Now(),
}
kafkaManager.PublishOrderEvent(orderEvent)
```

**Consumer**: `kafka/manager.go:consumeOrderEvents()`
```go
// Process updated orders
go func() {
    for {
        msg, err := updatedReader.ReadMessage(ctx)
        // Send notification to user based on status
        var message string
        switch event.Status {
        case "confirmed":
            message = "Đơn hàng của bạn đã được xác nhận"
        case "shipped":
            message = "Đơn hàng của bạn đã được giao vận"
        // ...
        }
        userNotification := NotificationEvent{
            UserID:    event.UserID,
            Type:      "order_updated",
            Message:   message,
            OrderID:   event.OrderID,
        }
        m.PublishNotification(userNotification)
    }
}()
```

### 3. `admin-notifications`
**Mục đích**: Thông báo dành cho admin (đơn hàng mới, etc.)

**Producer**: Kafka Consumer tự động tạo từ `order-created` events
**Consumer**: `kafka/manager.go:consumeAdminNotifications()`

### 4. `user-notifications`  
**Mục đích**: Thông báo dành cho user (trạng thái đơn hàng)

**Producer**: Kafka Consumer tự động tạo từ `order-updated` events
**Consumer**: `kafka/manager.go:consumeUserNotifications()`

## 🔧 Kafka Configuration

### Connection Setup
```go
// kafka/manager.go
func NewManager(brokers []string) *Manager {
    m := &Manager{
        brokers: brokers,
        writers: make(map[string]*kafka.Writer),
        readers: make(map[string]*kafka.Reader),
    }
    
    // Initialize writers for each topic
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
```

### Consumer Groups
```go
// Consumer for order created events
createdReader := kafka.NewReader(kafka.ReaderConfig{
    Brokers:  m.brokers,
    Topic:    TopicOrderCreated,
    GroupID:  "order-processor",        // Consumer group
    MinBytes: 1,                        // Min bytes to read
    MaxBytes: 10e6,                     // Max bytes to read
})
```

## 📨 Event Structures

### OrderEvent
```go
type OrderEvent struct {
    OrderID     uint      `json:"order_id"`
    UserID      uint      `json:"user_id"`
    Type        string    `json:"type"`     // "created", "updated"
    Status      string    `json:"status"`
    TotalAmount float64   `json:"total_amount"`
    Timestamp   time.Time `json:"timestamp"`
}
```

### NotificationEvent
```go
type NotificationEvent struct {
    UserID    uint      `json:"user_id,omitempty"`
    Type      string    `json:"type"`     // "order_created", "order_updated", "admin_notification"
    Message   string    `json:"message"`
    OrderID   uint      `json:"order_id,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}
```

## 🔄 Event Flow

### 1. User đặt hàng
```
POST /api/orders 
    ↓
CreateOrder() 
    ↓
┌─ Save to Database
│
├─ Publish "order-created" event
│       ↓
│   Kafka Consumer nhận event
│       ↓
│   Publish "admin-notification" 
│       ↓
│   Admin WebSocket broadcast
│
└─ Direct WebSocket broadcast (backup)
```

### 2. Admin cập nhật trạng thái
```
PUT /api/admin/orders/:id
    ↓
UpdateOrderStatus()
    ↓
┌─ Update Database  
│
├─ Publish "order-updated" event
│       ↓
│   Kafka Consumer nhận event
│       ↓
│   Publish "user-notification"
│       ↓
│   User WebSocket broadcast
│
└─ Direct WebSocket broadcast (backup)
```

## 🚀 Cách chạy và debug

### 1. Khởi động Kafka
```bash
docker-compose up -d
```

### 2. Kiểm tra Kafka Topics
```bash
# List topics
docker exec kafka-demo-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Read messages từ topic
docker exec kafka-demo-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic order-created \
  --from-beginning
```

### 3. Monitor Consumer Groups
```bash
# List consumer groups
docker exec kafka-demo-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 --list

# Check consumer group status  
docker exec kafka-demo-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group order-processor --describe
```

### 4. Reset Consumer Offset (nếu cần)
```bash
docker exec kafka-demo-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group order-processor \
  --reset-offsets --to-earliest --all-topics --execute
```

## 🐛 Troubleshooting

### 1. Consumer không nhận message
**Triệu chứng**: Logs hiển thị "EOF" errors
**Giải pháp**:
```bash
# Reset consumer group
docker exec kafka-demo-kafka kafka-consumer-groups \
  --bootstrap-server localhost:9092 \
  --group order-processor \
  --reset-offsets --to-latest --all-topics --execute
  
# Restart backend
cd backend && go run main.go
```

### 2. Topic không tồn tại
**Triệu chứng**: "Topic not found" errors
**Giải pháp**:
```bash
# Create topic manually
docker exec kafka-demo-kafka kafka-topics \
  --create --topic order-created \
  --bootstrap-server localhost:9092 \
  --partitions 1 --replication-factor 1
```

### 3. Kafka timeout
**Triệu chứng**: "context deadline exceeded"
**Giải pháp**: Tăng timeout trong consumer config:
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
```

## 📊 Monitoring với Kafdrop

Truy cập `http://localhost:9000` để:
- Xem tất cả topics và partitions
- Monitor consumer groups và lag
- Đọc messages trong topics
- Debug Kafka cluster health

## 🎯 Best Practices

1. **Error Handling**: Luôn handle errors trong consumers, không để crash
2. **Timeout**: Set reasonable timeout cho consumers để tránh block
3. **Consumer Groups**: Sử dụng consumer groups cho scalability
4. **Message Format**: Consistent JSON format cho tất cả events
5. **Backup Mechanisms**: WebSocket broadcast trực tiếp như fallback
6. **Monitoring**: Sử dụng Kafdrop để monitor Kafka health

## 🔗 Dependencies

```go
go.mod:
- github.com/segmentio/kafka-go  // Kafka client
- github.com/gin-gonic/gin       // HTTP framework
- gorm.io/gorm                   // ORM
- github.com/golang-jwt/jwt/v5   // JWT
```

Kafka implementation này đảm bảo **reliability**, **scalability** và **real-time performance** cho hệ thống đặt hàng! 🚀"# kafka" 
