#!/bin/bash

echo "🚀 Starting Microservices..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Build and start all services
echo "📦 Building and starting microservices with Docker Compose..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Health check
echo "🔍 Checking service health..."

echo "Checking API Gateway (port 8080)..."
curl -f http://localhost:8080/api/products > /dev/null 2>&1 && echo "✅ API Gateway is ready" || echo "❌ API Gateway is not ready"

echo "Checking Auth Service (port 8081)..."
curl -f http://localhost:8081/auth/login > /dev/null 2>&1 && echo "✅ Auth Service is ready" || echo "❌ Auth Service is not ready"

echo "Checking Product Service (port 8082)..."
curl -f http://localhost:8082/api/products > /dev/null 2>&1 && echo "✅ Product Service is ready" || echo "❌ Product Service is not ready"

echo "Checking Order Service (port 8083)..."
curl -f http://localhost:8083/api/orders > /dev/null 2>&1 && echo "✅ Order Service is ready" || echo "❌ Order Service is not ready"

echo ""
echo "🎉 Microservices are starting up!"
echo "📊 Services running on:"
echo "  🚪 API Gateway: http://localhost:8080"
echo "  🔐 Auth Service: http://localhost:8081"
echo "  📦 Product Service: http://localhost:8082"
echo "  🛍️ Order Service: http://localhost:8083"
echo ""
echo "📚 View logs: docker-compose logs -f"
echo "🛑 Stop services: docker-compose down"