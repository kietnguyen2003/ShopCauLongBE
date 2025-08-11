# 🛒 Kafka Order Demo - Backend

Backend service cho hệ thống quản lý đơn hàng cầu lông với Kafka event streaming và WebSocket real-time.

## 🏗️ Kiến trúc (Domain Driven Design + Clean Architecture)

```
backend/
├── cmd/
│   └── server/
│       └── main.go                 # 🚀 Entry point của ứng dụng
├── internal/
│   ├── domain/                     # 📋 Domain Layer (Business Logic)
│   │   ├── auth/
│   │   │   └── user.go            # User entity và business rules
│   │   ├── order/
│   │   │   └── order.go           # Order entity và business logic
│   │   └── product/
│   │       └── product.go         # Product entity và validation
│   ├── application/                # 🔄 Application Layer (Use Cases)
│   │   ├── auth/
│   │   │   └── service.go         # Authentication use cases
│   │   ├── order/
│   │   │   └── service.go         # Order management use cases
│   │   └── product/
│   │       └── service.go         # Product management use cases
│   ├── infrastructure/             # 🔧 Infrastructure Layer (External)
│   │   ├── config/
│   │   │   └── config.go          # Configuration management
│   │   ├── database/
│   │   │   ├── database.go        # Database connection & seeding
│   │   │   ├── gorm_user_repository.go
│   │   │   ├── gorm_product_repository.go
│   │   │   └── gorm_order_repository.go
│   │   ├── kafka/
│   │   │   └── event_bus.go       # Kafka event publishing
│   │   └── websocket/
│   │       └── hub.go             # WebSocket management
│   └── interfaces/                 # 🌐 Interface Layer (Presentation)
│       └── http/
│           ├── auth_handler.go     # Authentication endpoints
│           ├── order_handler.go    # Order management endpoints
│           └── product_handler.go  # Product endpoints
├── go.mod
├── go.sum
├── Dockerfile                      # 🐳 Container configuration
└── README.md
```

## 🎯 Nguyên tắc Clean Architecture

### 1. **Domain Layer** 📋
- **Trách nhiệm**: Business logic, entities, domain rules
- **Đặc điểm**: Độc lập hoàn toàn, không phụ thuộc layer nào khác
- **Chứa**: User, Order, Product entities với validation

### 2. **Application Layer** 🔄
- **Trách nhiệm**: Use cases, orchestration, business workflows  
- **Đặc điểm**: Sử dụng domain entities, định nghĩa interfaces
- **Chứa**: Services cho Auth, Order, Product management

### 3. **Infrastructure Layer** 🔧
- **Trách nhiệm**: External concerns (database, messaging, websockets)
- **Đặc điểm**: Implement interfaces được định nghĩa trong application layer
- **Chứa**: Database repos, Kafka event bus, WebSocket hub

### 4. **Interface Layer** 🌐
- **Trách nhiệm**: HTTP endpoints, request/response handling
- **Đặc điểm**: Chuyển đổi HTTP requests thành application use cases
- **Chứa**: REST API handlers

## 🚀 Công nghệ sử dụng

- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL + GORM
- **Message Queue**: Apache Kafka (IBM Sarama)
- **Real-time**: WebSocket (Gorilla)
- **Authentication**: JWT
- **Containerization**: Docker

## 📦 Cài đặt & Chạy

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL
- Apache Kafka

### Development
```bash
# 1. Clone repository
git clone <repo-url>
cd backend

# 2. Install dependencies
go mod tidy

# 3. Run with Docker Compose (recommended)
docker-compose up -d

# 4. Or run locally
go run cmd/server/main.go
```

### Build
```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Build Docker image
docker build -t kafka-order-backend .
```

## 🔧 Cấu hình

Cấu hình thông qua environment variables:

```bash
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/kafka_demo?sslmode=disable
JWT_SECRET=your-secret-key
KAFKA_BROKERS=localhost:9092
```

## 📡 API Endpoints

### Authentication
- `POST /auth/register` - Đăng ký user
- `POST /auth/login` - Đăng nhập user
- `POST /auth/admin-login` - Đăng nhập admin

### Products
- `GET /api/products` - Lấy danh sách sản phẩm
- `GET /api/products/:id` - Lấy chi tiết sản phẩm

### Orders (Protected)
- `POST /api/orders` - Tạo đơn hàng
- `GET /api/orders` - Lấy đơn hàng của user

### Admin (Protected + Admin Role)
- `GET /api/admin/orders` - Lấy tất cả đơn hàng
- `PUT /api/admin/orders/:id` - Cập nhật trạng thái đơn hàng

### WebSocket
- `GET /ws/user/:userID` - WebSocket cho user
- `GET /ws/admin` - WebSocket cho admin

## 🎯 Event Flow

1. **Tạo đơn hàng** → Kafka event `order-created` → WebSocket notification cho admin
2. **Cập nhật đơn hàng** → Kafka event `order-updated` → WebSocket notification cho user

## 🏗️ DDD Components

### Domain Entities
```go
// User entity với business rules
type User struct {
    ID       uint
    Username string
    Email    string
    // Business methods
    IsValidForLogin() bool
    UpdatePassword(newPassword string) error
}

// Order aggregate với business logic
type Order struct {
    ID          uint
    UserID      uint
    OrderItems  []OrderItem
    Status      OrderStatus
    // Business methods
    AddItem(productID uint, quantity int) error
    UpdateStatus(status OrderStatus) error
    CanBeModified() bool
}
```

### Repository Interfaces
```go
// Định nghĩa trong domain layer
type UserRepository interface {
    Create(user *User) error
    GetByID(id uint) (*User, error)
    GetByUsername(username string) (*User, error)
}

// Implement trong infrastructure layer
type GormUserRepository struct {
    db *gorm.DB
}
```

### Use Cases
```go
// Application layer services
type AuthService struct {
    userRepo  auth.UserRepository
    jwtSecret string
}

func (s *AuthService) Login(req LoginRequest) (*AuthResponse, error) {
    // Business logic orchestration
    user, err := s.userRepo.GetByUsername(req.Username)
    // Validation, JWT generation, etc.
}
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Test specific layer
go test ./internal/domain/...
go test ./internal/application/...
```

## 🔄 Kafka Event Flow

```
Order Created → Kafka Producer → order-created topic → Consumer → WebSocket Admin
Order Updated → Kafka Producer → order-updated topic → Consumer → WebSocket User
```

## 🐛 Troubleshooting

### Build Issues
```bash
# Clean module cache
go clean -modcache
go mod download
go mod tidy
```

### Database Issues
```bash
# Reset database
docker-compose down -v
docker-compose up -d
```

## 📄 License

MIT License - xem file LICENSE để biết thêm chi tiết.

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Liên hệ

- **Developer**: [Your Name]
- **Email**: [your-email@example.com]
- **Project Link**: [https://github.com/your-username/kafka-order-demo](https://github.com/your-username/kafka-order-demo) 
