package services

import (
	"context"
	"log"
	"time"

	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/google/uuid"
)

type TimerService struct {
	db          *database.PostgresDB
	alertEngine *AlertEngine
}

func NewTimerService(db *database.PostgresDB, ae *AlertEngine) *TimerService {
	ts := &TimerService{
		db:          db,
		alertEngine: ae,
	}
	// Start background worker to check for expired timers
	go ts.expirationWorker()
	return ts
}

func (s *TimerService) StartTimer(ctx context.Context, userID uuid.UUID, durationSecs int, label string, lat, lng float64) (*models.Timer, error) {
	expiresAt := time.Now().Add(time.Duration(durationSecs) * time.Second)
	
	timer := &models.Timer{
		ID:              uuid.New(),
		UserID:          userID,
		DurationSeconds: durationSecs,
		Label:           label,
		Lat:             lat,
		Lng:             lng,
		ExpiresAt:       expiresAt,
		Status:          models.TimerStatusActive,
		CreatedAt:       time.Now(),
	}

	if err := s.db.CreateTimer(ctx, timer); err != nil {
		return nil, err
	}
	return timer, nil
}

func (s *TimerService) ExtendTimer(ctx context.Context, timerID uuid.UUID, additionalSeconds int) error {
	return s.db.ExtendTimer(ctx, timerID, additionalSeconds)
}

func (s *TimerService) MarkSafe(ctx context.Context, timerID uuid.UUID) error {
	return s.db.UpdateTimerStatus(ctx, timerID, models.TimerStatusSafe)
}

// Background goroutine that checks for expired active timers every 10 seconds
func (s *TimerService) expirationWorker() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		ctx := context.Background()
		expiredTimers, err := s.db.GetExpiredActiveTimers(ctx)
		if err != nil {
			log.Printf("Error checking for expired timers: %v", err)
			continue
		}

		for _, timer := range expiredTimers {
			// Mark as EXPIRED immediately so it's only processed once
			if err := s.db.UpdateTimerStatus(ctx, timer.ID, models.TimerStatusExpired); err != nil {
				log.Printf("Error marking timer %s expired: %v", timer.ID, err)
				continue
			}

			// Trigger SOS alert for user
			user, err := s.db.GetUserByID(ctx, timer.UserID)
			if err != nil || user == nil {
				log.Printf("Error fetching user for expired timer %s", timer.ID)
				continue
			}

			hb := &models.Heartbeat{
				UserID:    user.ID,
				Lat:       timer.Lat,
				Lng:       timer.Lng,
				Timestamp: time.Now(),
				AccuracyM: 10,
			}

			// Send Alert to contacts
			err = s.alertEngine.SendAlertToContacts(
				ctx,
				user,
				hb,
				100, // Highest severity
				"TIMER EXPIRED: " + timer.Label,
			)
			if err != nil {
				log.Printf("Failed to send alert for expired timer %s: %v", timer.ID, err)
			} else {
				log.Printf("Successfully triggered SOS alert for expired timer %s", timer.ID)
			}
		}
	}
}
