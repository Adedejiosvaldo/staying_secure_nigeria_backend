package services

import (
	"context"
	"time"

	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/google/uuid"
)

type IncidentService struct {
	db *database.PostgresDB
}

func NewIncidentService(db *database.PostgresDB) *IncidentService {
	return &IncidentService{
		db: db,
	}
}

func (s *IncidentService) ReportIncident(ctx context.Context, userID uuid.UUID, hazardType, desc string, lat, lng float64) (*models.Incident, error) {
	incident := &models.Incident{
		ID:          uuid.New(),
		UserID:      userID,
		HazardType:  hazardType,
		Description: desc,
		Lat:         lat,
		Lng:         lng,
		CreatedAt:   time.Now(),
	}

	if err := s.db.CreateIncident(ctx, incident); err != nil {
		return nil, err
	}

	// In a complete implementation, this might broadcast to nearby users or connected websocket clients

	return incident, nil
}
