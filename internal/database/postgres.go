package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	pool *pgxpool.Pool
}

func NewPostgresDB(databaseURL string) (*PostgresDB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &PostgresDB{pool: pool}, nil
}

func (db *PostgresDB) Close() {
	db.pool.Close()
}

// User operations
func (db *PostgresDB) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, phone, name, trusted_contacts, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.pool.Exec(ctx, query,
		user.ID, user.Phone, user.Name, user.TrustedContacts,
		user.Settings, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (db *PostgresDB) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, phone, name, trusted_contacts, settings, created_at, updated_at
		FROM users WHERE id = $1
	`
	var user models.User
	err := db.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Phone, &user.Name, &user.TrustedContacts,
		&user.Settings, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *PostgresDB) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	query := `
		SELECT id, phone, name, trusted_contacts, settings, created_at, updated_at
		FROM users WHERE phone = $1
	`
	var user models.User
	err := db.pool.QueryRow(ctx, query, phone).Scan(
		&user.ID, &user.Phone, &user.Name, &user.TrustedContacts,
		&user.Settings, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *PostgresDB) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET name = $2, trusted_contacts = $3, settings = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := db.pool.Exec(ctx, query,
		user.ID, user.Name, user.TrustedContacts, user.Settings, time.Now(),
	)
	return err
}

// Heartbeat operations
func (db *PostgresDB) CreateHeartbeat(ctx context.Context, hb *models.Heartbeat) error {
	query := `
		INSERT INTO heartbeats (id, user_id, source, lat, lng, accuracy_m, cell_info, battery_pct, speed, last_gasp, timestamp, signature, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := db.pool.Exec(ctx, query,
		hb.ID, hb.UserID, hb.Source, hb.Lat, hb.Lng, hb.AccuracyM,
		hb.CellInfo, hb.BatteryPct, hb.Speed, hb.LastGasp, hb.Timestamp,
		hb.Signature, hb.CreatedAt,
	)
	return err
}

func (db *PostgresDB) GetLatestHeartbeat(ctx context.Context, userID uuid.UUID) (*models.Heartbeat, error) {
	query := `
		SELECT id, user_id, source, lat, lng, accuracy_m, cell_info, battery_pct, speed, last_gasp, timestamp, signature, created_at
		FROM heartbeats
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`
	var hb models.Heartbeat
	err := db.pool.QueryRow(ctx, query, userID).Scan(
		&hb.ID, &hb.UserID, &hb.Source, &hb.Lat, &hb.Lng, &hb.AccuracyM,
		&hb.CellInfo, &hb.BatteryPct, &hb.Speed, &hb.LastGasp, &hb.Timestamp,
		&hb.Signature, &hb.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &hb, nil
}

func (db *PostgresDB) GetHeartbeatsSince(ctx context.Context, userID uuid.UUID, since time.Time) ([]models.Heartbeat, error) {
	query := `
		SELECT id, user_id, source, lat, lng, accuracy_m, cell_info, battery_pct, speed, last_gasp, timestamp, signature, created_at
		FROM heartbeats
		WHERE user_id = $1 AND timestamp >= $2
		ORDER BY timestamp DESC
	`
	rows, err := db.pool.Query(ctx, query, userID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var heartbeats []models.Heartbeat
	for rows.Next() {
		var hb models.Heartbeat
		err := rows.Scan(
			&hb.ID, &hb.UserID, &hb.Source, &hb.Lat, &hb.Lng, &hb.AccuracyM,
			&hb.CellInfo, &hb.BatteryPct, &hb.Speed, &hb.LastGasp, &hb.Timestamp,
			&hb.Signature, &hb.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		heartbeats = append(heartbeats, hb)
	}
	return heartbeats, nil
}

// LastGasp operations
func (db *PostgresDB) CreateLastGasp(ctx context.Context, lg *models.LastGasp) error {
	query := `
		INSERT INTO last_gasps (id, user_id, lat, lng, accuracy_m, cell_info, created_at, expiry_ts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := db.pool.Exec(ctx, query,
		lg.ID, lg.UserID, lg.Lat, lg.Lng, lg.AccuracyM,
		lg.CellInfo, lg.CreatedAt, lg.ExpiryTs,
	)
	return err
}

func (db *PostgresDB) GetActiveLastGasp(ctx context.Context, userID uuid.UUID) (*models.LastGasp, error) {
	query := `
		SELECT id, user_id, lat, lng, accuracy_m, cell_info, created_at, expiry_ts
		FROM last_gasps
		WHERE user_id = $1 AND expiry_ts > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`
	var lg models.LastGasp
	err := db.pool.QueryRow(ctx, query, userID).Scan(
		&lg.ID, &lg.UserID, &lg.Lat, &lg.Lng, &lg.AccuracyM,
		&lg.CellInfo, &lg.CreatedAt, &lg.ExpiryTs,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lg, nil
}

// Alert operations
func (db *PostgresDB) CreateAlert(ctx context.Context, alert *models.Alert) error {
	query := `
		INSERT INTO alerts (id, user_id, state, score, reason, sent_to, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	sentToJSON, _ := models.StringArray(alert.SentTo).Value()
	_, err := db.pool.Exec(ctx, query,
		alert.ID, alert.UserID, alert.State, alert.Score, alert.Reason,
		sentToJSON, alert.CreatedAt,
	)
	return err
}

func (db *PostgresDB) GetLatestAlert(ctx context.Context, userID uuid.UUID) (*models.Alert, error) {
	query := `
		SELECT id, user_id, state, score, reason, sent_to, created_at, resolved_at
		FROM alerts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var alert models.Alert
	var sentTo models.StringArray
	err := db.pool.QueryRow(ctx, query, userID).Scan(
		&alert.ID, &alert.UserID, &alert.State, &alert.Score, &alert.Reason,
		&sentTo, &alert.CreatedAt, &alert.ResolvedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	alert.SentTo = sentTo
	return &alert, nil
}

func (db *PostgresDB) ResolveAlert(ctx context.Context, alertID uuid.UUID) error {
	query := `UPDATE alerts SET resolved_at = NOW() WHERE id = $1`
	_, err := db.pool.Exec(ctx, query, alertID)
	return err
}

// Blackbox operations
func (db *PostgresDB) CreateBlackboxTrail(ctx context.Context, trail *models.BlackboxTrail) error {
	query := `
		INSERT INTO blackbox_trails (id, user_id, start_ts, end_ts, data_points, file_url, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := db.pool.Exec(ctx, query,
		trail.ID, trail.UserID, trail.StartTs, trail.EndTs,
		trail.DataPoints, trail.FileURL, trail.UploadedAt,
	)
	return err
}

func (db *PostgresDB) GetBlackboxTrails(ctx context.Context, userID uuid.UUID, limit int) ([]models.BlackboxTrail, error) {
	query := `
		SELECT id, user_id, start_ts, end_ts, data_points, file_url, uploaded_at
		FROM blackbox_trails
		WHERE user_id = $1
		ORDER BY uploaded_at DESC
		LIMIT $2
	`
	rows, err := db.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trails []models.BlackboxTrail
	for rows.Next() {
		var trail models.BlackboxTrail
		err := rows.Scan(
			&trail.ID, &trail.UserID, &trail.StartTs, &trail.EndTs,
			&trail.DataPoints, &trail.FileURL, &trail.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		trails = append(trails, trail)
	}
	return trails, nil
}

// Contact management operations
func (db *PostgresDB) AddContact(ctx context.Context, userID uuid.UUID, contact map[string]string) error {
	// Convert map to Contact struct
	newContact := models.Contact{
		ID:    contact["id"],
		Name:  contact["name"],
		Phone: contact["phone"],
	}

	query := `
		UPDATE users
		SET trusted_contacts = trusted_contacts || $1::jsonb,
			updated_at = NOW()
		WHERE id = $2
	`
	contactJSON, err := json.Marshal([]models.Contact{newContact})
	if err != nil {
		return err
	}

	_, err = db.pool.Exec(ctx, query, contactJSON, userID)
	return err
}

func (db *PostgresDB) UpdateContact(ctx context.Context, userID uuid.UUID, contactID string, updates map[string]string) error {
	// Get current contacts
	user, err := db.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Update the contact in the array
	updated := false
	for i, contact := range user.TrustedContacts {
		if contact.ID == contactID {
			if name, ok := updates["name"]; ok && name != "" {
				user.TrustedContacts[i].Name = name
			}
			if phone, ok := updates["phone"]; ok && phone != "" {
				user.TrustedContacts[i].Phone = phone
			}
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("contact not found")
	}

	// Save back to database
	contactsJSON, err := json.Marshal(user.TrustedContacts)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET trusted_contacts = $1::jsonb,
			updated_at = NOW()
		WHERE id = $2
	`
	_, err = db.pool.Exec(ctx, query, contactsJSON, userID)
	return err
}

func (db *PostgresDB) DeleteContact(ctx context.Context, userID uuid.UUID, contactID string) error {
	// Get current contacts
	user, err := db.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Filter out the contact to delete
	newContacts := make([]models.Contact, 0)
	found := false
	for _, contact := range user.TrustedContacts {
		if contact.ID != contactID {
			newContacts = append(newContacts, contact)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("contact not found")
	}

	// Save back to database
	contactsJSON, err := json.Marshal(newContacts)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET trusted_contacts = $1::jsonb,
			updated_at = NOW()
		WHERE id = $2
	`
	_, err = db.pool.Exec(ctx, query, contactsJSON, userID)
	return err
}
