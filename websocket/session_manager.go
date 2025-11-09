package websocket

import (
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gofiber/websocket/v2"
)

type Session struct {
	UserID        primitive.ObjectID
	Username      string
	Conn          *websocket.Conn
	ConnectedAt   time.Time
	LastAccrualAt time.Time
	LastHeartbeat time.Time
	IsActive      bool
}

type SessionManager struct {
	sessions map[primitive.ObjectID]*Session
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[primitive.ObjectID]*Session),
	}
}

func (sm *SessionManager) AddSession(userID primitive.ObjectID, username string, conn *websocket.Conn) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session := &Session{
		UserID:        userID,
		Username:      username,
		Conn:          conn,
		ConnectedAt:   time.Now(),
		LastAccrualAt: time.Now(),
		LastHeartbeat: time.Now(),
		IsActive:      true,
	}

	sm.sessions[userID] = session
	log.Printf("✅ Session created for user: %s (%s)", username, userID.Hex())
	return session
}

func (sm *SessionManager) GetSession(userID primitive.ObjectID) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	return session, exists
}

func (sm *SessionManager) RemoveSession(userID primitive.ObjectID) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.IsActive = false
		delete(sm.sessions, userID)
		log.Printf("🗑️ Session removed for user: %s (%s)", session.Username, userID.Hex())
	}
}

func (sm *SessionManager) UpdateHeartbeat(userID primitive.ObjectID) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.LastHeartbeat = time.Now()
	}
}

func (sm *SessionManager) UpdateLastAccrual(userID primitive.ObjectID) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.LastAccrualAt = time.Now()
	}
}

func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *SessionManager) GetActiveSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	activeSessions := make([]*Session, 0)
	for _, session := range sm.sessions {
		if session.IsActive {
			activeSessions = append(activeSessions, session)
		}
	}
	return activeSessions
}

func (sm *SessionManager) CheckInactiveSessions(timeout time.Duration) []primitive.ObjectID {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	inactiveUsers := make([]primitive.ObjectID, 0)

	for userID, session := range sm.sessions {
		if session.IsActive && now.Sub(session.LastHeartbeat) > timeout {
			session.IsActive = false
			inactiveUsers = append(inactiveUsers, userID)
			log.Printf("⚠️ Session marked inactive due to heartbeat timeout: %s (%s)", 
				session.Username, userID.Hex())
		}
	}

	return inactiveUsers
}