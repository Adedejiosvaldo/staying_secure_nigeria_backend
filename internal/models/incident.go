package models

import (
	"time"
	"github.com/google/uuid"
)

// Incident represents a user-reported hazard or emergency on the map
type Incident struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	HazardType  string    `json:"hazard_type" db:"hazard_type"`
	Description string    `json:"description" db:"description"`
	Lat         float64   `json:"lat" db:"lat"`
	Lng         float64   `json:"lng" db:"lng"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
