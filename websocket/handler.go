package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/golang-jwt/jwt/v4"

	"go-ubipay-websocket/config"
	"go-ubipay-websocket/database"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebSocketHandler struct {
	cfg            *config.Config
	sessionManager *SessionManager
	db             *database.Database
}

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type HeartbeatMessage struct {
	Timestamp int64 `json:"timestamp"`
}

func NewWebSocketHandler(cfg *config.Config, sessionManager *SessionManager, db *database.Database) *WebSocketHandler {
	return &WebSocketHandler{
		cfg:            cfg,
		sessionManager: sessionManager,
		db:             db,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (h *WebSocketHandler) WebSocketConnection(c *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("⚠️ WebSocket connection panic: %v", err)
		}
	}()

	// Extract token from query parameters
	token := c.Query("token")
	var userID primitive.ObjectID
	var username string
	fmt.Println("token:",token)

	if token != "" {
		// Validate JWT token
		authenticatedUserID, authenticatedUsername, err := h.validateSessionToken(token)
		if err != nil {
			log.Printf("❌ JWT validation failed: %v", err)
			c.WriteJSON(WSMessage{
				Type:    "auth_failed",
				Payload: "Invalid or expired token",
			})
			c.Close()
			return
		}
		userID = authenticatedUserID
		username = authenticatedUsername
		log.Printf("🔌 WebSocket connection established for authenticated user: %s (%s)", username, userID.Hex())
	} else {
		// Use mock user data for testing when no token provided
		userID, _ = primitive.ObjectIDFromHex("507f1f77bcf86cd799439011")
		username = "testuser@example.com"
		log.Printf("🔌 WebSocket connection established for test user: %s (%s)", username, userID.Hex())
	}

	// Add session to manager
	session := h.sessionManager.AddSession(userID, username, c)
	defer h.sessionManager.RemoveSession(userID)

	// Ensure user wallet exists in database
	_, err := h.db.GetUserWallet(userID)
	if err != nil {
		// Create wallet if it doesn't exist
		_, err := h.db.CreateUserWallet(userID)
		if err != nil {
			log.Printf("❌ Failed to create wallet for user %s: %v", username, err)
		}
	}

	// Send initial connection success message
	c.WriteJSON(WSMessage{
		Type:    "connected",
		Payload: fiber.Map{"user_id": userID.Hex(), "username": username},
	})

	// Start heartbeat ticker
	heartbeatTicker := time.NewTicker(h.cfg.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	// Message handling loop
	for {
		select {
		case <-heartbeatTicker.C:
			// Send heartbeat ping
			err := c.WriteJSON(WSMessage{
				Type:    "heartbeat",
				Payload: HeartbeatMessage{Timestamp: time.Now().Unix()},
			})
			if err != nil {
				log.Printf("❌ Failed to send heartbeat to user %s: %v", username, err)
				return
			}

		default:
			// Read incoming messages
			messageType, msg, err := c.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("❌ WebSocket read error for user %s: %v", username, err)
				}
				return
			}

			if messageType == websocket.TextMessage {
				h.handleMessage(session, msg)
			}
		}
	}
}

func (h *WebSocketHandler) handleMessage(session *Session, msg []byte) {
	var wsMsg WSMessage
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		log.Printf("❌ Failed to parse WebSocket message from user %s: %v", session.Username, err)
		return
	}

	switch wsMsg.Type {
	case "heartbeat":
		h.sessionManager.UpdateHeartbeat(session.UserID)
		log.Printf("💓 Heartbeat received from user: %s", session.Username)

	case "auth":
		h.handleAuthMessage(session, wsMsg.Payload)

	case "balance_request":
		h.handleBalanceRequest(session)

	default:
		log.Printf("⚠️ Unknown message type from user %s: %s", session.Username, wsMsg.Type)
	}
}

func (h *WebSocketHandler) handleBalanceRequest(session *Session) {
	wallet, err := h.db.GetUserWallet(session.UserID)
	if err != nil {
		log.Printf("❌ Failed to get wallet for user %s: %v", session.Username, err)
		session.Conn.WriteJSON(WSMessage{
			Type:    "error",
			Payload: "Failed to retrieve balance",
		})
		return
	}
	log.Printf("💳 Balance request for user %s - Real balance: %d", session.Username, wallet.Balance)

	session.Conn.WriteJSON(WSMessage{
		Type:    "balance",
		Payload: fiber.Map{"balance": wallet.Balance},
	})

	log.Printf("💰 Balance sent to user: %s - %d points", session.Username, wallet.Balance)
}

func (h *WebSocketHandler) SendAccrualNotification(session *Session, points int, newBalance int) {
	err := session.Conn.WriteJSON(WSMessage{
		Type: "accrual",
		Payload: fiber.Map{
			"points":      points,
			"new_balance": newBalance,
			"timestamp":   time.Now().Unix(),
		},
	})
	if err != nil {
		log.Printf("❌ Failed to send accrual notification to user %s: %v", session.Username, err)
	} else {
		log.Printf("📢 Accrual notification sent to user %s: +%d points, new balance: %d",
			session.Username, points, newBalance)
	}
}

func (h *WebSocketHandler) SendBalanceUpdate(session *Session, balance int) {
	err := session.Conn.WriteJSON(WSMessage{
		Type: "balance_update",
		Payload: fiber.Map{
			"balance":   balance,
			"timestamp": time.Now().Unix(),
		},
	})
	if err != nil {
		log.Printf("❌ Failed to send balance update to user %s: %v", session.Username, err)
	} else {
		log.Printf("💳 Balance update sent to user %s: %d points", session.Username, balance)
	}
}

func (h *WebSocketHandler) validateSessionToken(sessionToken string) (primitive.ObjectID, string, error) {
	// Use database to validate session token
	user, err := h.db.GetUserBySessionToken(sessionToken)
	if err != nil {
		log.Printf("❌ Session token validation failed: %v", err)
		return primitive.NilObjectID, "", err
	}

	// Check if user is enabled
	if !user.Enable {
		log.Printf("❌ User account is disabled: %s", user.Username)
		return primitive.NilObjectID, "", jwt.ErrInvalidKey
	}

	// Return user information
	return user.ID, user.Username, nil
}

func (h *WebSocketHandler) handleAuthMessage(session *Session, payload interface{}) {
	// Extract token from payload
	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		session.Conn.WriteJSON(WSMessage{
			Type:    "auth_failed",
			Payload: "Invalid auth message format",
		})
		return
	}

	token, ok := payloadMap["token"].(string)
	if !ok || token == "" {
		session.Conn.WriteJSON(WSMessage{
			Type:    "auth_failed",
			Payload: "Token is required",
		})
		return
	}

	// Validate JWT token
	userID, username, err := h.validateSessionToken(token)
	if err != nil {
		log.Printf("❌ Auth message validation failed: %v", err)
		session.Conn.WriteJSON(WSMessage{
			Type:    "auth_failed",
			Payload: "Invalid or expired token",
		})
		return
	}

	// Update session with authenticated user data
	session.UserID = userID
	session.Username = username
	log.Printf("✅ Authentication successful for user: %s (%s)", username, userID.Hex())

	// Send authentication success message
	session.Conn.WriteJSON(WSMessage{
		Type:    "auth_success",
		Payload: fiber.Map{"user_id": userID.Hex(), "username": username},
	})
}
