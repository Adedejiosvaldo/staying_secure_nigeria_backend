package database

import (
	"context"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Trip operations
func (db *PostgresDB) CreateTrip(ctx context.Context, trip *models.Trip) error {
	query := `
		INSERT INTO trips (id, user_id, destination, estimated_arrival, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.pool.Exec(ctx, query,
		trip.ID, trip.UserID, trip.Destination, trip.EstimatedArrival, trip.IsActive, trip.CreatedAt,
	)
	return err
}

func (db *PostgresDB) GetActiveTrip(ctx context.Context, userID uuid.UUID) (*models.Trip, error) {
	query := `
		SELECT id, user_id, destination, estimated_arrival, is_active, created_at
		FROM trips
		WHERE user_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1
	`
	var trip models.Trip
	err := db.pool.QueryRow(ctx, query, userID).Scan(
		&trip.ID, &trip.UserID, &trip.Destination, &trip.EstimatedArrival, &trip.IsActive, &trip.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (db *PostgresDB) EndTrip(ctx context.Context, tripID uuid.UUID) error {
	query := `UPDATE trips SET is_active = false WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, tripID)
	return err
}

func (db *PostgresDB) AddTripLocation(ctx context.Context, loc *models.TripLocation) error {
	query := `
		INSERT INTO trip_locations (trip_id, lat, lng, timestamp)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.pool.Exec(ctx, query,
		loc.TripID, loc.Lat, loc.Lng, loc.Timestamp,
	)
	return err
}

func (db *PostgresDB) GetTripLocations(ctx context.Context, tripID uuid.UUID) ([]models.TripLocation, error) {
	query := `
		SELECT trip_id, lat, lng, timestamp
		FROM trip_locations
		WHERE trip_id = $1
		ORDER BY timestamp ASC
	`
	rows, err := db.pool.Query(ctx, query, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locs []models.TripLocation
	for rows.Next() {
		var loc models.TripLocation
		err := rows.Scan(&loc.TripID, &loc.Lat, &loc.Lng, &loc.Timestamp)
		if err != nil {
			return nil, err
		}
		locs = append(locs, loc)
	}
	return locs, nil
}
