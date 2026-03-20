package models

import (
	"time"
	"github.com/google/uuid"
)

// Trip represents a live trip session
type Trip struct {
	ID               uuid.UUID `json:"id" db:"id"`
	UserID           uuid.UUID `json:"user_id" db:"user_id"`
	Destination      string    `json:"destination" db:"destination"`
	EstimatedArrival time.Time `json:"estimated_arrival" db:"estimated_arrival"`
	IsActive         bool      `json:"is_active" db:"is_active"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// TripLocation represents a coordinate point during a trip
type TripLocation struct {
	TripID    uuid.UUID `json:"trip_id" db:"trip_id"`
	Lat       float64   `json:"lat" db:"lat"`
	Lng       float64   `json:"lng" db:"lng"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// TripGuardian represents a view of a trusted contact observing a trip
type TripGuardian struct {
	ContactID   string `json:"contact_id"`
	Name        string `json:"name"`
	LastViewed  time.Time `json:"last_viewed"`
}
