package models

import (
	"time"
	"github.com/google/uuid"
)

// Timer represents a Sentinel Mode countdown timer
type Timer struct {
	ID              uuid.UUID `json:"id" db:"id"`
	UserID          uuid.UUID `json:"user_id" db:"user_id"`
	DurationSeconds int       `json:"duration_seconds" db:"duration_seconds"`
	Label           string    `json:"label" db:"label"`
	Lat             float64   `json:"lat" db:"lat"`
	Lng             float64   `json:"lng" db:"lng"`
	ExpiresAt       time.Time `json:"expires_at" db:"expires_at"`
	Status          string    `json:"status" db:"status"` // "ACTIVE", "SAFE", "EXPIRED"
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

const (
	TimerStatusActive  = "ACTIVE"
	TimerStatusSafe    = "SAFE"
	TimerStatusExpired = "EXPIRED"
)
