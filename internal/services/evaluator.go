package services

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/adedejiosvaldo/safetrace/backend/internal/config"
	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
)

const (
	StateSafe         = "SAFE"
	StateCaution      = "CAUTION"
	StateAtRisk       = "AT_RISK"
	StateAlert        = "ALERT"
	StateWaitLastGasp = "WAIT_LASTGASP"
)

type SafetyEvaluator struct {
	cfg      *config.Config
	postgres *database.PostgresDB
	redis    *database.RedisDB
	alerter  *AlertEngine
}

func NewSafetyEvaluator(
	cfg *config.Config,
	postgres *database.PostgresDB,
	redis *database.RedisDB,
	alerter *AlertEngine,
) *SafetyEvaluator {
	return &SafetyEvaluator{
		cfg:      cfg,
		postgres: postgres,
		redis:    redis,
		alerter:  alerter,
	}
}

type EvaluationResult struct {
	State  string
	Score  int
	Reason string
}

// EvaluateUserSafety is the main entry point for safety evaluation
func (se *SafetyEvaluator) EvaluateUserSafety(ctx context.Context, userID uuid.UUID) (*EvaluationResult, error) {
	// Check for active LastGasp
	lastGasp, err := se.postgres.GetActiveLastGasp(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check lastgasp: %w", err)
	}

	if lastGasp != nil {
		// User has active LastGasp - wait period
		return &EvaluationResult{
			State:  StateWaitLastGasp,
			Score:  0,
			Reason: "LastGasp active - monitoring connectivity",
		}, nil
	}

	// Get latest heartbeat
	heartbeat, err := se.postgres.GetLatestHeartbeat(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get heartbeat: %w", err)
	}

	if heartbeat == nil {
		// No heartbeat data yet
		return &EvaluationResult{
			State:  StateSafe,
			Score:  100,
			Reason: "No heartbeat data yet",
		}, nil
	}

	// Run deterministic checks first
	deterministicResult := se.checkDeterministicRules(heartbeat)
	if deterministicResult != nil {
		return deterministicResult, nil
	}

	// Calculate composite score
	score := se.calculateSafetyScore(ctx, userID, heartbeat)

	// Map score to state
	var state string
	var reason string

	switch {
	case score >= 80:
		state = StateSafe
		reason = "All indicators normal"
	case score >= 50:
		state = StateCaution
		reason = "Some indicators concerning - silent check initiated"
	default:
		state = StateAtRisk
		reason = "Multiple risk indicators detected"
	}

	result := &EvaluationResult{
		State:  state,
		Score:  score,
		Reason: reason,
	}

	// Update state in Redis
	userState := &models.UserState{
		UserID:        userID,
		State:         state,
		Score:         score,
		LastHeartbeat: heartbeat.Timestamp,
		UpdatedAt:     time.Now(),
	}
	se.redis.SetUserState(ctx, userState)

	// Handle state transitions
	if err := se.handleStateTransition(ctx, userID, state, score, reason); err != nil {
		return nil, fmt.Errorf("failed to handle state transition: %w", err)
	}

	return result, nil
}

// checkDeterministicRules applies hard rules that override scoring
func (se *SafetyEvaluator) checkDeterministicRules(hb *models.Heartbeat) *EvaluationResult {
	// Rule 1: Recent heartbeat within window
	timeSinceHeartbeat := time.Since(hb.Timestamp)
	if timeSinceHeartbeat < time.Duration(se.cfg.HeartbeatWindowSeconds)*time.Second {
		if hb.LastGasp {
			// LastGasp received but recent - monitor
			return &EvaluationResult{
				State:  StateCaution,
				Score:  60,
				Reason: "LastGasp received - monitoring",
			}
		}
		// Normal recent heartbeat
		return nil // Continue to scoring
	}

	// Rule 2: Sudden stop detection (if speed data available)
	if hb.Speed != nil && *hb.Speed > 40 {
		// Check previous heartbeat for sudden deceleration
		// (This would require looking at previous heartbeat - simplified here)
		// If speed dropped from >40 to <5 in short time, immediate alert
	}

	// Rule 3: Heartbeat too old
	if timeSinceHeartbeat > time.Duration(se.cfg.HeartbeatWindowSeconds)*time.Second {
		missedMinutes := int(timeSinceHeartbeat.Minutes())
		return &EvaluationResult{
			State:  StateAtRisk,
			Score:  30,
			Reason: fmt.Sprintf("No heartbeat for %d minutes", missedMinutes),
		}
	}

	return nil
}

// calculateSafetyScore computes composite safety score (0-100)
func (se *SafetyEvaluator) calculateSafetyScore(ctx context.Context, userID uuid.UUID, hb *models.Heartbeat) int {
	score := 0

	// Component 1: Heartbeat recency (30 points)
	timeSinceHeartbeat := time.Since(hb.Timestamp)
	recencyMinutes := timeSinceHeartbeat.Minutes()
	
	switch {
	case recencyMinutes < 5:
		score += 30
	case recencyMinutes < 10:
		score += 20
	case recencyMinutes < 15:
		score += 10
	default:
		score += 0
	}

	// Component 2: GPS accuracy (20 points)
	switch {
	case hb.AccuracyM < 50:
		score += 20
	case hb.AccuracyM < 200:
		score += 15
	case hb.AccuracyM < 500:
		score += 10
	default:
		score += 5
	}

	// Component 3: Movement pattern (20 points)
	// Check if speed is consistent with expected behavior
	if hb.Speed != nil {
		speed := *hb.Speed
		switch {
		case speed >= 0 && speed < 100: // Normal speed
			score += 20
		case speed >= 100: // Unusually high speed
			score += 10
		}
	} else {
		score += 15 // No speed data, neutral
	}

	// Component 4: Signal quality (10 points)
	if hb.CellInfo.RSSI > -70 {
		score += 10
	} else if hb.CellInfo.RSSI > -90 {
		score += 5
	}

	// Component 5: Source reliability (5 points)
	if hb.Source == "http" {
		score += 5
	} else {
		score += 3 // SMS fallback
	}

	// Component 6: Battery level (15 points)
	if hb.BatteryPct != nil {
		switch {
		case *hb.BatteryPct > 20:
			score += 15
		case *hb.BatteryPct > 5:
			score += 10
		default:
			score += 5
		}
	} else {
		score += 10 // Unknown, neutral
	}

	// Ensure score is within bounds
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// handleStateTransition creates alerts and triggers notifications
func (se *SafetyEvaluator) handleStateTransition(ctx context.Context, userID uuid.UUID, newState string, score int, reason string) error {
	// Get previous state
	prevState, err := se.redis.GetUserState(ctx, userID)
	if err != nil {
		return err
	}

	// Only act on state changes or critical states
	if prevState != nil && prevState.State == newState && newState != StateAlert {
		return nil // No change, no action needed
	}

	// Check if alert was recently sent (deduplication)
	if newState == StateAtRisk || newState == StateAlert {
		alreadySent, err := se.redis.CheckAlertSent(ctx, userID, 5*time.Minute)
		if err != nil {
			return err
		}
		if alreadySent {
			return nil // Don't spam alerts
		}
	}

	// Handle state-specific actions
	switch newState {
	case StateCaution:
		// Silent ping - would trigger FCM notification to user
		// Implemented in alert engine
		return nil

	case StateAtRisk, StateAlert:
		// Create alert record
		alert := &models.Alert{
			ID:        uuid.New(),
			UserID:    userID,
			State:     models.AlertState(newState),
			Score:     score,
			Reason:    reason,
			SentTo:    []string{},
			CreatedAt: time.Now(),
		}

		if err := se.postgres.CreateAlert(ctx, alert); err != nil {
			return fmt.Errorf("failed to create alert: %w", err)
		}

		// Get user details for notification
		user, err := se.postgres.GetUserByID(ctx, userID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if user == nil {
			return fmt.Errorf("user not found: %s", userID)
		}

		// Get latest heartbeat for location
		hb, err := se.postgres.GetLatestHeartbeat(ctx, userID)
		if err != nil {
			return err
		}

		// Send alerts to trusted contacts
		go func() {
			ctx := context.Background()
			if err := se.alerter.SendAlertToContacts(ctx, user, hb, score, reason); err != nil {
				// Log error (in production, use proper logging)
				fmt.Printf("Failed to send alerts: %v\n", err)
			}

			// Mark alert as sent
			se.redis.MarkAlertSent(ctx, userID, 5*time.Minute)
		}()
	}

	return nil
}

// DetectSuddenStop checks for sudden deceleration between heartbeats
func (se *SafetyEvaluator) DetectSuddenStop(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Get last 2 heartbeats
	since := time.Now().Add(-5 * time.Minute)
	heartbeats, err := se.postgres.GetHeartbeatsSince(ctx, userID, since)
	if err != nil || len(heartbeats) < 2 {
		return false, err
	}

	latest := heartbeats[0]
	previous := heartbeats[1]

	// Check if both have speed data
	if latest.Speed == nil || previous.Speed == nil {
		return false, nil
	}

	// Detect sudden stop: speed dropped from >40 to <5 km/h
	if *previous.Speed > 40 && *latest.Speed < 5 {
		timeDiff := latest.Timestamp.Sub(previous.Timestamp).Seconds()
		if timeDiff < 60 { // Within 60 seconds
			// Calculate deceleration
			deceleration := (*previous.Speed - *latest.Speed) / 3.6 / timeDiff // m/s²
			if deceleration > 6 { // > 6 m/s² is concerning
				return true, nil
			}
		}
	}

	return false, nil
}

// DetectTowerJump checks for suspicious cell tower changes
func (se *SafetyEvaluator) DetectTowerJump(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Get last 2 heartbeats
	since := time.Now().Add(-5 * time.Minute)
	heartbeats, err := se.postgres.GetHeartbeatsSince(ctx, userID, since)
	if err != nil || len(heartbeats) < 2 {
		return false, err
	}

	latest := heartbeats[0]
	previous := heartbeats[1]

	// Check if cell IDs are different
	if latest.CellInfo.CID == previous.CellInfo.CID {
		return false, nil
	}

	// Calculate distance between locations
	distance := haversineDistance(
		previous.Lat, previous.Lng,
		latest.Lat, latest.Lng,
	)

	timeDiff := latest.Timestamp.Sub(previous.Timestamp).Minutes()

	// If moved > 5km in < 2 minutes, suspicious
	if distance > 5.0 && timeDiff < 2 {
		return true, nil
	}

	return false, nil
}

// haversineDistance calculates distance between two GPS coordinates in km
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in km

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
