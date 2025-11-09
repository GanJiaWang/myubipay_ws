# Go UbiPay WebSocket Server - Setup Instructions (No Auth)

## Prerequisites

- Go 1.21 or higher
- MongoDB (local installation or MongoDB Atlas)
- Git

## Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd go-ubipay-websocket
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   ```
   
   Edit the `.env` file with your configuration:
   ```env
   SERVER_PORT=3000
   MONGODB_URI=mongodb://localhost:27017
   MONGODB_NAME=ubipay
   JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
   ACCRUAL_INTERVAL=1m
   POINTS_PER_MINUTE=1
   HEARTBEAT_INTERVAL=30s
   ```

## MongoDB Setup

### Option 1: Local MongoDB
1. Install MongoDB locally
2. Start MongoDB service
3. Create database named `ubipay`

### Option 2: MongoDB Atlas
1. Create a free account at [MongoDB Atlas](https://www.mongodb.com/atlas)
2. Create a cluster and get the connection string
3. Update `MONGODB_URI` in your `.env` file

## Running the Application

### Development Mode
```bash
go run main.go
```

### Build and Run
```bash
go build -o ubipay-websocket
./ubipay-websocket
```

### With Environment Variables
```bash
export $(cat .env | xargs) && go run main.go
```

## Testing the API

### 1. Connect via WebSocket (No Auth Required)
Use any WebSocket client - authentication has been disabled for testing:
```bash
# Using wscat (no auth headers needed)
wscat -c "ws://localhost:3000/ws"

# Or using the Node.js test script
node test-websocket.js
```

### 2. Check Health Status
```bash
curl http://localhost:3000/health
```

### 3. View Active Sessions
```bash
curl http://localhost:3000/admin/sessions
```

### 4. Trigger Manual Accrual (for testing)
```bash
curl -X POST http://localhost:3000/admin/accrual/run
```

## API Endpoints

- `GET /health` - Health check endpoint
- `GET /ws` - WebSocket connection (no authentication required)
- `POST /admin/accrual/run` - Trigger manual point accrual
- `GET /admin/sessions` - View active WebSocket sessions

## WebSocket Message Types

### Incoming Messages (Client → Server)
- `heartbeat` - Respond to server heartbeat
- `balance_request` - Request current balance

### Outgoing Messages (Server → Client)
- `connected` - Connection established
- `heartbeat` - Periodic ping
- `balance` - Current balance response
- `accrual` - Point accrual notification
- `error` - Error messages

## Configuration Options

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| SERVER_PORT | 3000 | HTTP server port |
| MONGODB_URI | mongodb://localhost:27017 | MongoDB connection string |
| MONGODB_NAME | ubipay | MongoDB database name |
| JWT_SECRET | (not used) | JWT signing secret (auth disabled) |
| ACCRUAL_INTERVAL | 1m | How often to run point accrual |
| POINTS_PER_MINUTE | 1 | Points awarded per minute |
| HEARTBEAT_INTERVAL | 30s | WebSocket heartbeat interval |

## Development

### Project Structure
```
go-ubipay-websocket/
├── main.go              # Application entry point
├── config/              # Configuration handling
├── models/              # Data models
├── database/            # MongoDB operations
├── websocket/           # WebSocket handlers
├── cron/                # Point accrual scheduler
├── middleware/            # (Authentication disabled for testing)
├── go.mod              # Go module dependencies
└── .env.example        # Environment variables template
```

### Adding New Features
1. Create new models in `models/`
2. Add database operations in `database/`
3. Implement handlers in appropriate packages
4. Update WebSocket message types if needed

## Troubleshooting

### Common Issues

1. **MongoDB Connection Failed**
   - Check if MongoDB is running
   - Verify connection string in `.env`

2. **WebSocket Connection Failed**
   2. Authentication is disabled - no tokens needed
   3. WebSocket connections are open without authentication

3. **Port Already in Use**
   - Change `SERVER_PORT` in `.env`
   - Kill existing process on the port

### Logs
The application logs all major events to console, including:
- WebSocket connections/disconnections
- Point accrual events
- Database operations
- ✅ WebSocket connections (no auth required)

## Production Deployment

### Environment Variables for Production
- Enable authentication by uncommenting auth middleware code
- Set proper `MONGODB_URI` for production database
- Configure appropriate timeouts and intervals

### Security Considerations
- Enable authentication for production use
- Use HTTPS in production
- Implement rate limiting
- Add proper CORS configuration
- Use environment-specific configurations

## Support

For issues and questions, please check:
- Application logs for error details
- MongoDB connection status
- WebSocket client implementation