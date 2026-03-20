package database

import (
	"context"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
)

// Incident operations
func (db *PostgresDB) CreateIncident(ctx context.Context, incident *models.Incident) error {
	query := `
		INSERT INTO incidents (id, user_id, hazard_type, description, lat, lng, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.pool.Exec(ctx, query,
		incident.ID, incident.UserID, incident.HazardType, incident.Description,
		incident.Lat, incident.Lng, incident.CreatedAt,
	)
	return err
}

// Get recent incidents in bounding box (for map view)
func (db *PostgresDB) GetRecentIncidents(ctx context.Context, limit int) ([]models.Incident, error) {
	query := `
		SELECT id, user_id, hazard_type, description, lat, lng, created_at
		FROM incidents
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := db.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []models.Incident
	for rows.Next() {
		var inc models.Incident
		err := rows.Scan(
			&inc.ID, &inc.UserID, &inc.HazardType, &inc.Description,
			&inc.Lat, &inc.Lng, &inc.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}
