package handlers

import (
	"net/http"
	"time"

	"github.com/adedejiosvaldo/safetrace/backend/internal/config"
	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/adedejiosvaldo/safetrace/backend/internal/services"
	"github.com/adedejiosvaldo/safetrace/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SMSHandler struct {
	cfg       *config.Config
	postgres  *database.PostgresDB
	redis     *database.RedisDB
	evaluator *services.SafetyEvaluator
	smsParser *services.SMSParser
}

func NewSMSHandler(
	cfg *config.Config,
	postgres *database.PostgresDB,
	redis *database.RedisDB,
	evaluator *services.SafetyEvaluator,
) *SMSHandler {
	return &SMSHandler{
		cfg:       cfg,
		postgres:  postgres,
		redis:     redis,
		evaluator: evaluator,
		smsParser: services.NewSMSParser(),
	}
}

// POST /v1/sms/webhook
// Twilio sends SMS data as form-encoded
func (h *SMSHandler) HandleIncomingSMS(c *gin.Context) {
	// Twilio sends data as form parameters
	body := c.PostForm("Body")

	if body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty message body"})
		return
	}

	// Parse SMS heartbeat
	heartbeat, err := h.smsParser.ParseHeartbeatSMS(body)
	if err != nil {
		// Log error and return success to Twilio to avoid retries
		c.XML(http.StatusOK, gin.H{"Response": "Message received but could not be parsed"})
		return
	}

	// Verify signature
	if !utils.VerifyStringSignature(
		body[:len(body)-len(heartbeat.Signature)-5], // Remove ";sig=..." part
		heartbeat.Signature,
		h.cfg.HMACSecret,
	) {
		c.XML(http.StatusOK, gin.H{"Response": "Invalid signature"})
		return
	}

	// Verify user exists
	user, err := h.postgres.GetUserByID(c.Request.Context(), heartbeat.UserID)
	if err != nil || user == nil {
		c.XML(http.StatusOK, gin.H{"Response": "User not found"})
		return
	}

	// Set metadata
	heartbeat.ID = uuid.New()
	heartbeat.Source = "sms"
	heartbeat.CreatedAt = time.Now()

	// Store heartbeat
	if err := h.postgres.CreateHeartbeat(c.Request.Context(), heartbeat); err != nil {
		c.XML(http.StatusOK, gin.H{"Response": "Storage error"})
		return
	}

	// Handle LastGasp if present
	if heartbeat.LastGasp {
		lastGasp := &models.LastGasp{
			ID:        uuid.New(),
			UserID:    heartbeat.UserID,
			Lat:       heartbeat.Lat,
			Lng:       heartbeat.Lng,
			AccuracyM: heartbeat.AccuracyM,
			CellInfo:  heartbeat.CellInfo,
			CreatedAt: time.Now(),
			ExpiryTs:  time.Now().Add(time.Duration(h.cfg.LastGaspTimeoutSeconds) * time.Second),
		}
		h.postgres.CreateLastGasp(c.Request.Context(), lastGasp)
	}

	// Trigger safety evaluation (async)
	go func() {
		ctx := c.Copy().Request.Context()
		h.evaluator.EvaluateUserSafety(ctx, heartbeat.UserID)
	}()

	// Respond with TwiML (Twilio expects this format)
	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?><Response><Message>Heartbeat received</Message></Response>`)
}
