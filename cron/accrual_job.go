package cron

import (
	"log"
	"time"

	"go-ubipay-websocket/config"
	"go-ubipay-websocket/database"
	"go-ubipay-websocket/websocket"

	"github.com/robfig/cron/v3"
)

type AccrualJob struct {
	cfg            *config.Config
	sessionManager *websocket.SessionManager
	db             *database.Database
	wsHandler      *websocket.WebSocketHandler
	cron           *cron.Cron
}

func NewAccrualJob(cfg *config.Config, sessionManager *websocket.SessionManager, db *database.Database, wsHandler *websocket.WebSocketHandler) *AccrualJob {
	return &AccrualJob{
		cfg:            cfg,
		sessionManager: sessionManager,
		db:             db,
		wsHandler:      wsHandler,
		cron:           cron.New(),
	}
}

func (j *AccrualJob) Start() {
	// Schedule accrual job to run every minute
	_, err := j.cron.AddFunc("@every 1m", j.runAccrual)
	if err != nil {
		log.Fatalf("❌ Failed to schedule accrual job: %v", err)
	}

	j.cron.Start()
	log.Println("✅ Accrual cron job started - running every minute")
}

func (j *AccrualJob) Stop() {
	j.cron.Stop()
	log.Println("🛑 Accrual cron job stopped")
}

func (j *AccrualJob) runAccrual() {
	startTime := time.Now()
	log.Printf("⏰ Starting accrual process at %s", startTime.Format("2006-01-02 15:04:05"))

	// Get all active sessions
	activeSessions := j.sessionManager.GetActiveSessions()
	log.Printf("📊 Found %d active sessions", len(activeSessions))

	if len(activeSessions) == 0 {
		log.Println("ℹ️ No active sessions found, skipping accrual")
		return
	}

	successCount := 0
	failureCount := 0

	for _, session := range activeSessions {
		if !session.IsActive {
			continue
		}

		// Calculate points to award (1 point per minute by default)
		pointsToAward := j.cfg.PointsPerMinute

		// Accrue points for this user
		err := j.db.AccruePoints(session.UserID, session.Username, pointsToAward)
		if err != nil {
			log.Printf("❌ Failed to accrue points for user %s: %v", session.Username, err)
			failureCount++
			continue
		}

		// Update last accrual time
		j.sessionManager.UpdateLastAccrual(session.UserID)

		// Get updated balance to send notification
		wallet, err := j.db.GetUserWallet(session.UserID)
		if err != nil {
			log.Printf("⚠️ Failed to get updated balance for user %s: %v", session.Username, err)
		} else {
			// Send real-time accrual notification via WebSocket
			wsSession, exists := j.sessionManager.GetSession(session.UserID)
			if exists && wsSession.IsActive {
				j.wsHandler.SendAccrualNotification(wsSession, pointsToAward, wallet.Balance)
			}
			log.Printf("💰 Accrued %d points for user %s, new balance: %d", 
				pointsToAward, session.Username, wallet.Balance)
		}

		successCount++
	}

	duration := time.Since(startTime)
	log.Printf("✅ Accrual process completed in %v - Success: %d, Failures: %d", 
		duration, successCount, failureCount)
}

func (j *AccrualJob) RunManualAccrual() {
	log.Println("🔧 Running manual accrual job")
	j.runAccrual()
}