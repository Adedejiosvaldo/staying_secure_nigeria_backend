package database

import (
	"context"
	"time"

	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Timer operations
func (db *PostgresDB) CreateTimer(ctx context.Context, timer *models.Timer) error {
	query := `
		INSERT INTO timers (id, user_id, duration_seconds, label, lat, lng, expires_at, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := db.pool.Exec(ctx, query,
		timer.ID, timer.UserID, timer.DurationSeconds, timer.Label,
		timer.Lat, timer.Lng, timer.ExpiresAt, timer.Status, timer.CreatedAt,
	)
	return err
}

func (db *PostgresDB) GetActiveTimer(ctx context.Context, userID uuid.UUID) (*models.Timer, error) {
	query := `
		SELECT id, user_id, duration_seconds, label, lat, lng, expires_at, status, created_at
		FROM timers
		WHERE user_id = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	var timer models.Timer
	err := db.pool.QueryRow(ctx, query, userID, models.TimerStatusActive).Scan(
		&timer.ID, &timer.UserID, &timer.DurationSeconds, &timer.Label,
		&timer.Lat, &timer.Lng, &timer.ExpiresAt, &timer.Status, &timer.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &timer, nil
}

func (db *PostgresDB) ExtendTimer(ctx context.Context, timerID uuid.UUID, additionalSeconds int) error {
	query := `
		UPDATE timers 
		SET duration_seconds = duration_seconds + $2,
		    expires_at = expires_at + make_interval(secs => $2)
		WHERE id = $1 AND status = $3
	`
	_, err := db.pool.Exec(ctx, query, timerID, additionalSeconds, models.TimerStatusActive)
	return err
}

func (db *PostgresDB) UpdateTimerStatus(ctx context.Context, timerID uuid.UUID, status string) error {
	query := `UPDATE timers SET status = $2 WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, timerID, status)
	return err
}

func (db *PostgresDB) GetExpiredActiveTimers(ctx context.Context) ([]models.Timer, error) {
	query := `
		SELECT id, user_id, duration_seconds, label, lat, lng, expires_at, status, created_at
		FROM timers
		WHERE status = $1 AND expires_at <= $2
	`
	rows, err := db.pool.Query(ctx, query, models.TimerStatusActive, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timers []models.Timer
	for rows.Next() {
		var timer models.Timer
		err := rows.Scan(
			&timer.ID, &timer.UserID, &timer.DurationSeconds, &timer.Label,
			&timer.Lat, &timer.Lng, &timer.ExpiresAt, &timer.Status, &timer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		timers = append(timers, timer)
	}
	return timers, nil
}
