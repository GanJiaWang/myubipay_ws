const WebSocket = require('ws');

// Configuration
const SERVER_URL = 'ws://localhost:3000/ws';

// Note: Authentication has been removed for testing
// WebSocket connections no longer require JWT tokens

class WebSocketTester {
    constructor() {
        this.ws = null;
        this.isConnected = false;
        this.heartbeatCount = 0;
        this.accrualCount = 0;
    }

    connect() {
        console.log('🔌 Connecting to WebSocket server...');
        
        this.ws = new WebSocket(SERVER_URL);

        this.ws.on('open', () => {
            console.log('✅ WebSocket connection established');
            this.isConnected = true;
        });

        this.ws.on('message', (data) => {
            try {
                const message = JSON.parse(data);
                this.handleMessage(message);
            } catch (error) {
                console.error('❌ Failed to parse message:', error);
            }
        });

        this.ws.on('close', (code, reason) => {
            console.log(`🔴 WebSocket connection closed: ${code} - ${reason}`);
            this.isConnected = false;
        });

        this.ws.on('error', (error) => {
            console.error('❌ WebSocket error:', error);
        });
    }

    handleMessage(message) {
        console.log('📨 Received message:', JSON.stringify(message, null, 2));

        switch (message.type) {
            case 'connected':
                console.log('🎉 Successfully authenticated and connected');
                console.log('User:', message.payload);
                break;

            case 'heartbeat':
                this.heartbeatCount++;
                console.log(`💓 Heartbeat #${this.heartbeatCount} received`);
                // Send heartbeat response
                this.sendHeartbeat();
                break;

            case 'balance':
                console.log(`💰 Current balance: ${message.payload.balance} points`);
                break;

            case 'accrual':
                this.accrualCount++;
                console.log(`🎯 Accrual #${this.accrualCount}: ${message.payload.points} points added`);
                console.log(`💳 New balance: ${message.payload.new_balance} points`);
                break;

            case 'error':
                console.error('🚨 Error from server:', message.payload);
                break;

            default:
                console.log('❓ Unknown message type:', message.type);
        }
    }

    sendHeartbeat() {
        const heartbeatMessage = {
            type: 'heartbeat',
            payload: {
                timestamp: Date.now()
            }
        };
        this.sendMessage(heartbeatMessage);
    }

    requestBalance() {
        const balanceMessage = {
            type: 'balance_request',
            payload: {}
        };
        console.log('📊 Requesting balance...');
        this.sendMessage(balanceMessage);
    }

    sendMessage(message) {
        if (this.ws && this.isConnected) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('❌ Cannot send message - WebSocket not connected');
        }
    }

    disconnect() {
        if (this.ws) {
            console.log('👋 Disconnecting from WebSocket...');
            this.ws.close();
        }
    }
}

// Main test function
async function runTest() {
    console.log('🧪 Starting WebSocket Test');
    console.log('='.repeat(50));

    const tester = new WebSocketTester();
    
    // Connect to WebSocket
    tester.connect();

    // Wait for connection to establish
    await new Promise(resolve => setTimeout(resolve, 2000));

    if (!tester.isConnected) {
        console.error('❌ Failed to establish WebSocket connection');
        return;
    }

    // Test balance request
    console.log('\n📊 Testing balance request...');
    tester.requestBalance();

    // Keep connection alive for 3 minutes to test accruals
    console.log('\n⏰ Keeping connection alive for 3 minutes to test point accruals...');
    console.log('💡 The server should award points every minute');
    console.log('💓 Heartbeats will be sent automatically');
    
    let minutes = 0;
    const interval = setInterval(() => {
        minutes++;
        console.log(`⏱️  ${minutes} minute(s) elapsed...`);
        
        if (minutes >= 3) {
            clearInterval(interval);
            console.log('\n✅ Test completed');
            console.log(`📊 Summary:`);
            console.log(`   - Heartbeats received: ${tester.heartbeatCount}`);
            console.log(`   - Accruals received: ${tester.accrualCount}`);
            tester.disconnect();
        }
    }, 60000);

    // Handle graceful shutdown
    process.on('SIGINT', () => {
        console.log('\n🛑 Received shutdown signal');
        clearInterval(interval);
        tester.disconnect();
        process.exit(0);
    });
}

// Check if WebSocket module is available
try {
    require('ws');
} catch (error) {
    console.error('❌ WebSocket module not found. Install it with:');
    console.error('   npm install ws');
    process.exit(1);
}

// Run the test
runTest().catch(console.error);
// ```

// To use this test script:

// 1. **Install the WebSocket module**:
//    ```bash
//    npm install ws
//    ```

// 2. **Start the Go server**:
//    ```bash
//    go run main.go
//    ```

// 2. **Run the test**:
//    ```bash
//    node test-websocket.js
//    ```

// The script will:
// - Connect to your WebSocket server
// - Handle authentication
// - Respond to heartbeats
// - Request balance information
// - Stay connected for 3 minutes to test point accruals
// - Display all WebSocket messages received

// This will help you verify that the WebSocket flow is working correctly without needing MongoDB.