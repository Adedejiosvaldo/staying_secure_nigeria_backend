package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// Database
	DatabaseURL string
	RedisURL    string

	// Security
	HMACSecret string
	JWTSecret  string

	// Twilio
	TwilioAccountSID  string
	TwilioAuthToken   string
	TwilioPhoneNumber string

	// Firebase
	FCMCredentialsPath string

	// Mapbox
	MapboxToken string

	// Thresholds
	HeartbeatIntervalSeconds int
	HeartbeatWindowSeconds   int
	LastGaspTimeoutSeconds   int
	SilentPromptSeconds      int
	BlackboxRetentionHours   int
}

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Port:                     getEnv("PORT", "8080"),
		DatabaseURL:              getEnv("DATABASE_URL", ""),
		RedisURL:                 getEnv("REDIS_URL", "redis://localhost:6379"),
		HMACSecret:               getEnv("HMAC_SECRET", ""),
		JWTSecret:                getEnv("JWT_SECRET", ""),
		TwilioAccountSID:         getEnv("TWILIO_ACCOUNT_SID", ""),
		TwilioAuthToken:          getEnv("TWILIO_AUTH_TOKEN", ""),
		TwilioPhoneNumber:        getEnv("TWILIO_PHONE_NUMBER", ""),
		FCMCredentialsPath:       getEnv("FCM_CREDENTIALS_PATH", ""),
		MapboxToken:              getEnv("MAPBOX_TOKEN", ""),
		HeartbeatIntervalSeconds: getEnvInt("HEARTBEAT_INTERVAL_SECONDS", 180),    // 3 min
		HeartbeatWindowSeconds:   getEnvInt("HEARTBEAT_WINDOW_SECONDS", 600),      // 10 min
		LastGaspTimeoutSeconds:   getEnvInt("LASTGASP_TIMEOUT_SECONDS", 3600),     // 60 min
		SilentPromptSeconds:      getEnvInt("SILENT_PROMPT_SECONDS", 10),          // 10 sec
		BlackboxRetentionHours:   getEnvInt("BLACKBOX_RETENTION_HOURS", 12),       // 12 hours
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.HMACSecret == "" {
		return fmt.Errorf("HMAC_SECRET is required")
	}
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.TwilioAccountSID == "" {
		return fmt.Errorf("TWILIO_ACCOUNT_SID is required")
	}
	if c.TwilioAuthToken == "" {
		return fmt.Errorf("TWILIO_AUTH_TOKEN is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
