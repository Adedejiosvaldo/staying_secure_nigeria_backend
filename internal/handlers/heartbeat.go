package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/adedejiosvaldo/safetrace/backend/internal/config"
	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/adedejiosvaldo/safetrace/backend/internal/services"
	"github.com/adedejiosvaldo/safetrace/backend/internal/utils"
)

type HeartbeatHandler struct {
	cfg       *config.Config
	postgres  *database.PostgresDB
	redis     *database.RedisDB
	evaluator *services.SafetyEvaluator
}

func NewHeartbeatHandler(
	cfg *config.Config,
	postgres *database.PostgresDB,
	redis *database.RedisDB,
	evaluator *services.SafetyEvaluator,
) *HeartbeatHandler {
	return &HeartbeatHandler{
		cfg:       cfg,
		postgres:  postgres,
		redis:     redis,
		evaluator: evaluator,
	}
}

type HeartbeatRequest struct {
	UserID     string           `json:"user_id" binding:"required"`
	Timestamp  time.Time        `json:"timestamp" binding:"required"`
	Lat        float64          `json:"lat" binding:"required"`
	Lng        float64          `json:"lng" binding:"required"`
	AccuracyM  int              `json:"accuracy_m" binding:"required"`
	CellInfo   models.CellInfo  `json:"cell_info" binding:"required"`
	BatteryPct *int             `json:"battery_pct,omitempty"`
	Speed      *float64         `json:"speed,omitempty"`
	LastGasp   bool             `json:"last_gasp"`
	Signature  string           `json:"signature" binding:"required"`
}

// POST /v1/heartbeat
func (h *HeartbeatHandler) CreateHeartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Rate limiting check
	allowed, err := h.redis.CheckRateLimit(c.Request.Context(), userID, 30*time.Second, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	// Verify user exists
	user, err := h.postgres.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Verify signature (excluding signature field itself)
	reqForVerification := map[string]interface{}{
		"user_id":     req.UserID,
		"timestamp":   req.Timestamp.Unix(),
		"lat":         req.Lat,
		"lng":         req.Lng,
		"accuracy_m":  req.AccuracyM,
		"cell_info":   req.CellInfo,
		"battery_pct": req.BatteryPct,
		"speed":       req.Speed,
		"last_gasp":   req.LastGasp,
	}

	if !utils.VerifySignature(reqForVerification, req.Signature, h.cfg.HMACSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

	// Create heartbeat record
	heartbeat := &models.Heartbeat{
		ID:         uuid.New(),
		UserID:     userID,
		Source:     "http",
		Lat:        req.Lat,
		Lng:        req.Lng,
		AccuracyM:  req.AccuracyM,
		CellInfo:   req.CellInfo,
		BatteryPct: req.BatteryPct,
		Speed:      req.Speed,
		LastGasp:   req.LastGasp,
		Timestamp:  req.Timestamp,
		Signature:  req.Signature,
		CreatedAt:  time.Now(),
	}

	// Store heartbeat
	if err := h.postgres.CreateHeartbeat(c.Request.Context(), heartbeat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store heartbeat"})
		return
	}

	// Handle LastGasp
	if req.LastGasp {
		lastGasp := &models.LastGasp{
			ID:        uuid.New(),
			UserID:    userID,
			Lat:       req.Lat,
			Lng:       req.Lng,
			AccuracyM: req.AccuracyM,
			CellInfo:  req.CellInfo,
			CreatedAt: time.Now(),
			ExpiryTs:  time.Now().Add(time.Duration(h.cfg.LastGaspTimeoutSeconds) * time.Second),
		}
		if err := h.postgres.CreateLastGasp(c.Request.Context(), lastGasp); err != nil {
			// Log error but don't fail the request
		}
	}

	// Trigger safety evaluation (async)
	go func() {
		ctx := c.Copy().Request.Context()
		if _, err := h.evaluator.EvaluateUserSafety(ctx, userID); err != nil {
			// Log error (in production, use proper logging)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "heartbeat received",
		"id":      heartbeat.ID,
	})
}

// GET /v1/user/:id/status
func (h *HeartbeatHandler) GetUserStatus(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Get user state from Redis
	state, err := h.redis.GetUserState(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get state"})
		return
	}

	if state == nil {
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"state":   "UNKNOWN",
			"message": "No data available",
		})
		return
	}

	c.JSON(http.StatusOK, state)
}

// POST /v1/alert/:id/resolve
func (h *HeartbeatHandler) ResolveAlert(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert_id"})
		return
	}

	if err := h.postgres.ResolveAlert(c.Request.Context(), alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "alert resolved",
	})
}
