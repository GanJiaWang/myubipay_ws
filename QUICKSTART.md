# Go UbiPay WebSocket Server - Quick Start Guide (No Auth)

## 🚀 Overview

A real-time point mining system built with Go Fiber that allows users to accrue points by maintaining active WebSocket connections. Features JWT authentication, MongoDB persistence, and automatic point accrual via cron jobs.

## 📋 Features

- ✅ WebSocket server (authentication disabled for testing)
- ✅ Real-time session tracking in memory
- ✅ Automatic point accrual every minute
- ✅ MongoDB integration for persistence
- ✅ Heartbeat verification for active connections
- ✅ Admin endpoints for monitoring
- ✅ Graceful shutdown handling

## 🛠️ Quick Setup

### 1. Prerequisites
```bash
# Install Go
brew install go  # macOS
# or download from https://golang.org/dl/

# Install MongoDB (optional for local development)
brew install mongodb/brew/mongodb-community
# or use MongoDB Atlas cloud
```

### 2. Clone and Setup
```bash
git clone <repository>
cd go-ubipay-websocket

# Install dependencies
go mod download

# Configure environment
cp .env.example .env
# Edit .env with your settings
```

### 3. Environment Configuration
```env
SERVER_PORT=3000
MONGODB_URI=mongodb://localhost:27017
MONGODB_NAME=ubipay
JWT_SECRET=your-super-secret-key-change-in-production
ACCRUAL_INTERVAL=1m
POINTS_PER_MINUTE=1
HEARTBEAT_INTERVAL=30s
```

### 4. Run the Application
```bash
# Development mode
go run main.go

# Or build and run
go build -o ubipay-websocket
./ubipay-websocket
```

## 🔌 API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/ws` | GET | WebSocket connection (no auth required) |
| `/admin/sessions` | GET | View active sessions |
| `/admin/accrual/run` | POST | Trigger manual accrual |

## 🔓 Authentication (Disabled for Testing)

### Connect via WebSocket (No Auth Required)
```bash
# Using wscat (no auth headers needed)
wscat -c "ws://localhost:3000/ws"

# Or programmatically in JavaScript
const ws = new WebSocket('ws://localhost:3000/ws');
```

**Note:** Authentication has been disabled for testing. All WebSocket connections are accepted without JWT tokens.

## 📊 MongoDB Collections

### TblUserWallet
```json
{
  "UserID": ObjectId,
  "WalletType": 1,
  "WalletName": "Point Wallet",
  "Balance": 1200,
  "Enable": true
}
```

### TblTransactionMovement
```json
{
  "UserID": ObjectId,
  "Username": "user@example.com",
  "TransactionType": 2,
  "TargetType": 1,
  "Amount": 300,
  "BeforeAmt": 0,
  "AfterAmt": 300
}
```

## 🎯 WebSocket Protocol

### Client → Server Messages
- `heartbeat` - Respond to server pings
- `balance_request` - Request current balance

### Server → Client Messages
- `connected` - Connection established
- `heartbeat` - Periodic ping (30s interval)
- `balance` - Current balance response
- `accrual` - Point accrual notification
- `error` - Error messages

## ⚙️ Configuration Options

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | 3000 | HTTP server port |
| `MONGODB_URI` | mongodb://localhost:27017 | MongoDB connection string |
| `MONGODB_NAME` | ubipay | Database name |
| `JWT_SECRET` | (not used) | JWT signing secret (auth disabled) |
| `ACCRUAL_INTERVAL` | 1m | Point accrual frequency |
| `POINTS_PER_MINUTE` | 1 | Points per minute of activity |
| `HEARTBEAT_INTERVAL` | 30s | WebSocket heartbeat frequency |

## 🚦 Monitoring

### Check Health
```bash
curl http://localhost:3000/health
```

### View Active Sessions
```bash
curl http://localhost:3000/admin/sessions
```

### Trigger Manual Accrual
```bash
curl -X POST http://localhost:3000/admin/accrual/run
```

## 🐛 Troubleshooting

### Common Issues

1. **MongoDB Connection Failed**
   - Check if MongoDB is running
   - Verify connection string in `.env`

2. **WebSocket Connection Failed**
   - Verify JWT token is valid
   - Check authorization header format

3. **Port Already in Use**
   - Change `SERVER_PORT` in `.env`
   - Kill existing process on the port

### Logs
The application logs all major events to console:
- ✅ WebSocket connections/disconnections
- ✅ Point accrual events
- ✅ Database operations
- ✅ Authentication events

## 🚀 Production Deployment

### Environment Variables
```env
JWT_SECRET=very-strong-random-secret-key
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/ubipay
```

### Security Considerations
- Use HTTPS in production
- Rotate JWT secrets regularly
- Implement rate limiting
- Add proper CORS configuration

## 📁 Project Structure
```
go-ubipay-websocket/
├── main.go                 # Application entry point
├── config/                 # Configuration handling
├── models/                 # Data models
├── database/               # MongoDB operations
├── websocket/             # WebSocket handlers
├── cron/                  # Point accrual scheduler
├── middleware/            # Authentication middleware
├── go.mod                # Dependencies
├── .env.example          # Environment template
└── SETUP.md              # Detailed setup guide
```

## 🆘 Support

For issues:
1. Check application logs
2. Verify MongoDB connection
3. Ensure proper JWT token format
4. Check WebSocket client implementation

## 📄 License

This project is part of the UbiPay Real-Time Point Mining System MVP.