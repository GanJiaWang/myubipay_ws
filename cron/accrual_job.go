package cron

import (
	"log"
	"time"

	"go-ubipay-websocket/config"
	"go-ubipay-websocket/database"
	"go-ubipay-websocket/websocket"

	"strconv"
	"strings"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func Decimal128ToInt(d primitive.Decimal128) int {
	s := d.String()              // 例如 "3.0"
	s = strings.Split(s, ".")[0] // 取整数部分
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func (j *AccrualJob) runAccrual() {
	startTime := time.Now()
	log.Printf("⏰ Starting accrual process at %s", startTime.Format("2006-01-02 15:04:05"))

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

		pointsToAward := j.cfg.PointsPerMinute

		// 给用户加积分
		err := j.db.AccruePoints(session.UserID, session.Username, pointsToAward)
		if err != nil {
			log.Printf("❌ Failed to accrue points for user %s: %v", session.Username, err)
			failureCount++
			continue
		}

		j.sessionManager.UpdateLastAccrual(session.UserID)

		// 获取最新钱包余额
		wallet, _ := j.db.GetUserWallet(session.UserID)
		if wallet == nil {
			log.Printf("⚠️ Failed to get updated balance for user %s", session.Username)
			continue
		}

		balance := Decimal128ToInt(wallet.Balance)

		// WebSocket 通知
		wsSession, exists := j.sessionManager.GetSession(session.UserID)
		if exists && wsSession.IsActive {
			j.wsHandler.SendAccrualNotification(wsSession, pointsToAward, balance)
		}

		log.Printf("💰 Accrued %d points for user %s, new balance: %d", pointsToAward, session.Username, balance)
		successCount++
	}

	duration := time.Since(startTime)
	log.Printf("✅ Accrual process completed in %v - Success: %d, Failures: %d", duration, successCount, failureCount)
}

func (j *AccrualJob) RunManualAccrual() {
	log.Println("🔧 Running manual accrual job")
	j.runAccrual()
}
