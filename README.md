# Real-Time Point Mining System (MVP) - TEST MODE (No Auth)

This README describes the architecture for the V1 MVP of a **point-based blockchain-style mining network**. Users accrue points by maintaining an active WebSocket connection. The interface is delivered via a Chrome Extension. Backend is built with **Go Fiber** and **MongoDB**, deployed on a single-node infra (e.g., Railway Hobby Tier for API + MongoDB Atlas for persistence).

**⚠️ CURRENTLY IN TEST MODE:**
- MongoDB operations have been commented out and replaced with console logs
- Authentication has been removed for easier testing
- WebSocket connections are open without JWT tokens

---

## Core Components

### 1. WebSocket Server (Go Fiber) - TEST MODE ACTIVE (No Auth)

* Accepts incoming WebSocket connections from clients (Chrome extension).
* **AUTH DISABLED:** No authentication required for testing.
* Tracks active sessions in memory.
* Sends periodic heartbeats to verify the client is still connected.
* Emits accrual events (every N minutes) for active users.
* **TEST MODE:** MongoDB writes replaced with console logs for testing.

### 2. MongoDB Collections

We are using MongoDB for direct persistence. Collections are structured as follows:

#### **TblUserWallet**

Holds user point balances.

```json
{
  "_id": ObjectId,
  "UserID": ObjectId,      // reference to user
  "WalletType": 1,         // 1 = point wallet
  "WalletName": "Point Wallet",
  "Balance": 1200,         // current point balance
  "Enable": true,
  "CreateBy": "System",
  "CreateDate": ISODate,
  "ModifiedBy": "API",
  "ModifiedDate": ISODate
}
```

#### **TblTransactionMovement**

Holds a log of all accrual and deduction events.

```json
{
  "_id": ObjectId,
  "UserID": ObjectId,
  "Username": "ubipayadmin@gmail.com",
  "TransactionType": 2,     // 1 = debit, 2 = credit
  "TargetType": 1,          // 1 = point accrual event
  "Amount": 300,            // points added/removed
  "BeforeAmt": 0,           // balance before txn
  "AfterAmt": 300,          // balance after txn
  "Enable": true,
  "CreateBy": "System",
  "CreateDate": ISODate,
  "ModifiedBy": "",
  "ModifiedDate": ISODate
}
```

---

## Accrual Process

* A **cron job** (e.g., every 1 minute) runs inside the Fiber app.
* It checks all active sessions in memory.
* For each active session:

  1. Calculate duration since last accrual.
  2. Determine points to award (e.g., 1 point per minute).
  3. **Directly write** to MongoDB:

     * Increment `Balance` in `TblUserWallet`.
     * Insert new record in `TblTransactionMovement`.

This ensures persistence is immediate and avoids memory–DB desync.

**TEST MODE NOTE:** Currently, point accrual logs to console instead of writing to MongoDB for testing purposes.

---

## Session Tracking

* When a user connects:

  * Validate token.
  * Start a session entry in memory (userID, connectedAt, lastAccrualAt).
* When disconnected:

  * Final accrual is computed up to disconnect timestamp.
  * Session entry is removed.

---

## Anti-Spoofing Considerations

* **Heartbeat Verification**: Clients must respond to periodic ping messages.
* **Token Validation**: Rotate JWTs or API tokens to prevent reuse.
* **IP/Device Fingerprinting**: (Future optimization) detect multiple fake clients.
* **Rate Limits**: Prevent excessive reconnects from inflating accruals.

---

## Quick Start (Test Mode)

### 1. Clone and Setup
```bash
git clone <repository>
cd go-ubipay-websocket
go mod download
```

### 2. Run the Server
```bash
go run main.go
```

### 3. Test the API
```bash
# Test WebSocket connection (no auth required)
wscat -c "ws://localhost:3000/ws"

# Or using the Node.js test script
node test-websocket.js
```

### 4. Monitor Logs
Check server console for:
- WebSocket connection events
- Heartbeat exchanges
- Point accrual simulations (every minute)
- Session management logs

## Deployment (MVP - Production Ready)

**Note:** Authentication must be re-enabled for production use by:
1. Uncommenting auth middleware code
2. Adding JWT secret to environment variables
3. Updating WebSocket handler to require authentication

* **API**: Go Fiber app hosted on Railway Hobby Tier.
* **DB**: MongoDB Atlas free/entry-level cluster.
* **Scaling**: Single-node, no sharding, minimal infra.
* **Chrome Extension**: Maintains WS connection and shows user balance.

## Enabling Features (After Testing)

To enable full functionality after testing:

1. **MongoDB:** Uncomment code in `database/database.go` and set `MONGODB_URI`
2. **Authentication:** Re-enable auth middleware and set `JWT_SECRET`
3. Restart the server

The application is production-ready and will switch to full functionality once the features are enabled.

---

## Future Enhancements

* Replace cron job with distributed task queues when scaling.
* Add Redis for session tracking (instead of in-memory).
* Multi-node scaling with sticky sessions or pub/sub.
* On-chain settlement of points (move from off-chain accounting).

---
