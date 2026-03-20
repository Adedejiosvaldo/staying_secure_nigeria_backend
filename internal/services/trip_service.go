package services

import (
	"context"
	"time"

	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/google/uuid"
)

type TripService struct {
	db *database.PostgresDB
}

func NewTripService(db *database.PostgresDB) *TripService {
	return &TripService{
		db: db,
	}
}

func (s *TripService) StartTrip(ctx context.Context, userID uuid.UUID, destination string, estArrival time.Time) (*models.Trip, error) {
	trip := &models.Trip{
		ID:               uuid.New(),
		UserID:           userID,
		Destination:      destination,
		EstimatedArrival: estArrival,
		IsActive:         true,
		CreatedAt:        time.Now(),
	}

	if err := s.db.CreateTrip(ctx, trip); err != nil {
		return nil, err
	}

	return trip, nil
}

func (s *TripService) EndTrip(ctx context.Context, tripID uuid.UUID) error {
	return s.db.EndTrip(ctx, tripID)
}

func (s *TripService) AddLocation(ctx context.Context, tripID uuid.UUID, lat, lng float64) error {
	loc := &models.TripLocation{
		TripID:    tripID,
		Lat:       lat,
		Lng:       lng,
		Timestamp: time.Now(),
	}
	return s.db.AddTripLocation(ctx, loc)
}

// In a real app we might use Redis to track who is currently viewing the trip via WebSockets
func (s *TripService) GetActiveGuardians(ctx context.Context, tripID uuid.UUID) ([]models.TripGuardian, error) {
	// Mock implementation
	return []models.TripGuardian{}, nil
}
