# 🛒 Order Management Demo - Microservices Architecture

Backend microservices cho hệ thống quản lý đơn hàng cầu lông.

## 🏗️ Kiến trúc Microservices

```
backend/
├── services/
│   ├── api-gateway/           # 🚪 API Gateway - Route requests
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── handlers/      # Middleware, Auth
│   │   │   └── proxy/         # Request proxying logic
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── auth-service/          # 🔐 Authentication Service
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── handlers/      # Auth endpoints
│   │   │   ├── models/        # User model
│   │   │   ├── repository/    # User repository
│   │   │   └── service/       # Auth business logic
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── product-service/       # 📦 Product Service
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── handlers/      # Product endpoints
│   │   │   ├── models/        # Product model
│   │   │   ├── repository/    # Product repository
│   │   │   └── service/       # Product business logic
│   │   ├── Dockerfile
│   │   └── go.mod
│   └── order-service/         # 🛍️ Order Service
│       ├── cmd/
│       │   └── main.go
│       ├── internal/
│       │   ├── config/
│       │   ├── handlers/      # Order endpoints
│       │   ├── models/        # Order & OrderItem models
│       │   ├── repository/    # Order repository
│       │   └── service/       # Order business logic
│       ├── Dockerfile
│       └── go.mod
└── docker-compose.yml
```

## 🎯 Microservices Design

### 1. **API Gateway** 🚪 (Port: 8080)
- **Trách nhiệm**: Route requests, Authentication middleware, CORS
- **Endpoints**: Proxy tất cả requests đến các services tương ứng
- **Đặc điểm**: Single entry point, handles auth validation

### 2. **Auth Service** 🔐 (Port: 8081)
- **Trách nhiệm**: User management, JWT authentication
- **Database**: auth_db (PostgreSQL)
- **Endpoints**:
  - `POST /auth/register` - Đăng ký user
  - `POST /auth/login` - Đăng nhập user
  - `POST /auth/admin-login` - Đăng nhập admin
  - `POST /auth/validate` - Validate JWT token

### 3. **Product Service** 📦 (Port: 8082)
- **Trách nhiệm**: Product catalog, inventory management
- **Database**: product_db (PostgreSQL)
- **Endpoints**:
  - `GET /api/products` - Lấy danh sách sản phẩm
  - `GET /api/products/:id` - Lấy chi tiết sản phẩm
  - `POST /api/products` - Tạo sản phẩm (Admin)
  - `PUT /api/products/:id/stock` - Cập nhật stock (Admin)
  - `POST /api/products/:id/decrease-stock` - Giảm stock (Internal)
  - `DELETE /api/products/:id` - Xóa sản phẩm (Admin)

### 4. **Order Service** 🛍️ (Port: 8083)
- **Trách nhiệm**: Order processing, order management
- **Database**: order_db (PostgreSQL)
- **Dependencies**: Calls Product Service for inventory
- **Endpoints**:
  - `POST /api/orders` - Tạo đơn hàng
  - `GET /api/orders` - Lấy đơn hàng của user
  - `GET /api/orders/:id` - Lấy chi tiết đơn hàng
  - `PUT /api/orders/:id` - Cập nhật trạng thái (Admin)
  - `GET /api/admin/orders` - Lấy tất cả đơn hàng (Admin)

## 🚀 Chạy Microservices

### Option 1: Docker Compose (Recommended)
```bash
# Chạy tất cả microservices với Docker
docker-compose up -d

# Xem logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Option 2: Chạy từng service riêng lẻ
```bash
# 1. Start databases
docker run -d --name auth-db -p 5433:5432 -e POSTGRES_DB=auth_db -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password postgres:15
docker run -d --name product-db -p 5434:5432 -e POSTGRES_DB=product_db -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password postgres:15
docker run -d --name order-db -p 5435:5432 -e POSTGRES_DB=order_db -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password postgres:15

# 2. Start services (in separate terminals)
cd services/auth-service && go run cmd/main.go
cd services/product-service && go run cmd/main.go  
cd services/order-service && go run cmd/main.go
cd services/api-gateway && go run cmd/main.go
```

## 🔧 Configuration

### Environment Variables

#### API Gateway
```bash
PORT=8080
AUTH_SERVICE_URL=http://localhost:8081
PRODUCT_SERVICE_URL=http://localhost:8082
ORDER_SERVICE_URL=http://localhost:8083
JWT_SECRET=your-secret-key
```

#### Auth Service
```bash
PORT=8081
DATABASE_URL=postgres://user:password@localhost:5433/auth_db?sslmode=disable
JWT_SECRET=your-secret-key
```

#### Product Service
```bash
PORT=8082
DATABASE_URL=postgres://user:password@localhost:5434/product_db?sslmode=disable
```

#### Order Service
```bash
PORT=8083
DATABASE_URL=postgres://user:password@localhost:5435/order_db?sslmode=disable
PRODUCT_SERVICE_URL=http://localhost:8082
```

## 📡 API Usage

### 1. Authentication
```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user1","email":"user1@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### 2. Products
```bash
# Get products
curl http://localhost:8080/api/products

# Get product by ID
curl http://localhost:8080/api/products/1
```

### 3. Orders (Requires Authentication)
```bash
# Create order
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "customer_name": "John Doe",
    "phone": "0123456789",
    "address": "123 Main St",
    "email": "john@example.com",
    "items": [
      {"product_id": 1, "quantity": 2}
    ]
  }'

# Get user orders
curl http://localhost:8080/api/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🏗️ Microservices Benefits

### ✅ Advantages
- **Scalability**: Mỗi service có thể scale độc lập
- **Technology Diversity**: Có thể dùng tech stack khác nhau cho mỗi service
- **Fault Isolation**: Lỗi ở 1 service không ảnh hưởng toàn bộ hệ thống
- **Team Independence**: Các team có thể develop và deploy độc lập
- **Database per Service**: Mỗi service có database riêng

### 🔄 Service Communication
- **Synchronous**: HTTP REST API calls giữa các services
- **API Gateway**: Centralized entry point và authentication
- **Service Discovery**: Static configuration (có thể nâng cấp lên Consul/Eureka)

## 🧪 Testing Services

### Health Check
```bash
# Check all services
curl http://localhost:8080/api/products  # Via Gateway
curl http://localhost:8081/auth/login    # Direct to Auth Service
curl http://localhost:8082/api/products  # Direct to Product Service
curl http://localhost:8083/api/orders    # Direct to Order Service (needs auth)
```

### Load Testing
```bash
# Install apache bench
apt-get install apache2-utils

# Test API Gateway
ab -n 1000 -c 10 http://localhost:8080/api/products
```

## 🐛 Troubleshooting

### Service Communication Issues
```bash
# Check network connectivity
docker network ls
docker network inspect backend_microservices-network

# Check service logs
docker-compose logs auth-service
docker-compose logs product-service
docker-compose logs order-service
docker-compose logs api-gateway
```

### Database Issues
```bash
# Connect to specific database
docker exec -it auth-db psql -U user -d auth_db
docker exec -it product-db psql -U user -d product_db  
docker exec -it order-db psql -U user -d order_db
```

### Port Conflicts
```bash
# Check port usage
netstat -tulpn | grep :8080
netstat -tulpn | grep :8081
netstat -tulpn | grep :8082
netstat -tulpn | grep :8083
```

## 🔄 Migration from Monolith

### Data Migration
1. Export data từ monolith database
2. Import vào các service databases tương ứng
3. Update foreign key references

### Gradual Migration Strategy
1. **Strangler Fig Pattern**: Gradually replace monolith endpoints
2. **Database per Service**: Migrate data từng domain
3. **API Versioning**: Maintain backward compatibility

## 📈 Monitoring & Observability

### Metrics
- Service health endpoints
- Database connection status
- Request/response times
- Error rates

### Logging
- Structured logging với correlation IDs
- Centralized log aggregation (ELK stack)
- Distributed tracing

## 🔒 Security

### Authentication Flow
1. Client → API Gateway → Auth Service (login)
2. Auth Service returns JWT token
3. Client → API Gateway (with JWT) → Protected Services
4. API Gateway validates JWT before forwarding

### Security Best Practices
- JWT với expiration time
- HTTPS for all external communication
- Service-to-service authentication (nếu cần)
- Input validation tại mỗi service

## 📄 License

MIT License - xem file LICENSE để biết thêm chi tiết.

## 🤝 Contributing

1. Fork the project
2. Create your feature branch (`git checkout -b feature/microservice-feature`)
3. Commit your changes (`git commit -m 'Add microservice feature'`)
4. Push to the branch (`git push origin feature/microservice-feature`)
5. Open a Pull Request

## 📞 Liên hệ

- **Developer**: [Your Name]
- **Email**: [your-email@example.com]
- **Project Link**: [https://github.com/your-username/kafka-order-demo](https://github.com/your-username/kafka-order-demo)