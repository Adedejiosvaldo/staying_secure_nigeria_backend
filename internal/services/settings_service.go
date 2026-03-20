package services

import (
	"context"
	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/google/uuid"
)

type SettingsService struct {
	db *database.PostgresDB
}

func NewSettingsService(db *database.PostgresDB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) UpdateUserSettings(ctx context.Context, userID uuid.UUID, newSettings models.UserSettings) error {
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	
	user.Settings = newSettings
	return s.db.UpdateUser(ctx, user)
}

func (s *SettingsService) PurgeUserData(ctx context.Context, userID uuid.UUID) error {
	return s.db.PurgeUserData(ctx, userID)
}
