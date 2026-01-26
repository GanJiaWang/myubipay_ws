package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort        string
	MongoDBURI        string
	MongoDBName       string
	JWTSecret         string
	AccrualInterval   time.Duration
	PointsPerMinute   int
	HeartbeatInterval time.Duration
}

func LoadConfig() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	return &Config{
		// ServerPort:        getEnv("SERVER_PORT", "3000"),
		// MongoDBURI:        os.Getenv("MONGODB_URI"),
		// MongoDBName:       os.Getenv("MONGODB_NAME"),
		// JWTSecret:         os.Getenv("JWT_SECRET"),
		ServerPort:        getEnv("SERVER_PORT", "3124"),
		MongoDBURI:        getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDBName:       getEnv("MONGODB_NAME", "ubipay"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		AccrualInterval:   getDurationEnv("ACCRUAL_INTERVAL", time.Minute),
		PointsPerMinute:   getIntEnv("POINTS_PER_MINUTE", 1),
		HeartbeatInterval: getDurationEnv("HEARTBEAT_INTERVAL", 30*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
