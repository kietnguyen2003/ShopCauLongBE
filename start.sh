#!/bin/bash

# Backend Development Script
echo "🛒 Starting Backend Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker Desktop first."
    exit 1
fi

echo "🐳 Docker is running. Starting database services..."

# Start database services
echo "🗄️ Starting PostgreSQL and pgAdmin..."
docker-compose up -d

# Wait for database to be ready
echo "⏳ Waiting for PostgreSQL to be ready..."
sleep 10

# Check service status
echo "📋 Database Service Status:"
docker-compose ps

echo ""
echo "✅ Database services started successfully!"
echo ""
echo "📡 Available Services:"
echo "   🗄️  PostgreSQL: localhost:5432"
echo "   🌐 pgAdmin:    http://localhost:5050"
echo "      - Email:    admin@example.com"
echo "      - Password: admin"
echo ""
echo "🚀 Now start the Go backend:"
echo "   go mod tidy"
echo "   go run cmd/server/main.go"
echo ""
echo "📊 To view database logs: docker-compose logs -f postgres"
echo "🛑 To stop database: docker-compose down"
echo ""
echo "Happy coding! 🎉"