#!/bin/bash

# Simple test script for Go UbiPay WebSocket Server
echo "🧪 Testing Go UbiPay WebSocket Server"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

# Check if MongoDB is running (optional)
if command -v mongosh &> /dev/null; then
    echo "✅ MongoDB client found"
else
    echo "⚠️  MongoDB client not found. Make sure MongoDB is running."
fi

# Build the application
echo "🔨 Building application..."
go build -o ubipay-websocket

if [ $? -eq 0 ]; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Creating from example..."
    cp .env.example .env
    echo "✅ Created .env file. Please edit it with your configuration."
fi

# Test basic functionality
echo "🧪 Running basic tests..."

# Test health endpoint (if server is running)
echo "🌐 Testing health endpoint..."
curl -s http://localhost:3000/health || echo "⚠️  Server not running on port 3000"

# Check dependencies
echo "📦 Checking dependencies..."
go mod verify

echo "✅ Test script completed"
echo ""
echo "To run the server:"
echo "  go run main.go"
echo ""
echo "To test with curl:"
echo "  curl http://localhost:3000/health"
echo ""
echo "To generate a test token:"
echo '  curl -X POST http://localhost:3000/auth/test-token \'
echo '    -H "Content-Type: application/json" \'
echo '    -d '\''{"user_id": "507f1f77bcf86cd799439011", "username": "testuser@example.com"}'\'''