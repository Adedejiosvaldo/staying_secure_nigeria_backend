package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/adedejiosvaldo/safetrace/backend/internal/config"
	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
)

type BlackboxHandler struct {
	cfg      *config.Config
	postgres *database.PostgresDB
}

func NewBlackboxHandler(
	cfg *config.Config,
	postgres *database.PostgresDB,
) *BlackboxHandler {
	return &BlackboxHandler{
		cfg:      cfg,
		postgres: postgres,
	}
}

type BlackboxUploadRequest struct {
	UserID     string                  `json:"user_id" binding:"required"`
	StartTs    time.Time               `json:"start_ts" binding:"required"`
	EndTs      time.Time               `json:"end_ts" binding:"required"`
	DataPoints []models.BlackboxEntry  `json:"data_points" binding:"required"`
}

// POST /v1/blackbox/upload
func (h *BlackboxHandler) UploadTrail(c *gin.Context) {
	var req BlackboxUploadRequest
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

	// Convert data points to JSON string (in production, store in S3/Spaces)
	dataJSON, err := json.Marshal(req.DataPoints)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize data"})
		return
	}

	// For now, store as data URI (in production, upload to object storage)
	fileURL := "data:application/json;base64," + string(dataJSON)

	// Create trail record
	trail := &models.BlackboxTrail{
		ID:         uuid.New(),
		UserID:     userID,
		StartTs:    req.StartTs,
		EndTs:      req.EndTs,
		DataPoints: len(req.DataPoints),
		FileURL:    fileURL,
		UploadedAt: time.Now(),
	}

	if err := h.postgres.CreateBlackboxTrail(c.Request.Context(), trail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store trail"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"trail_id":    trail.ID,
		"data_points": trail.DataPoints,
		"message":     "blackbox trail uploaded successfully",
	})
}

// GET /v1/blackbox/trails/:user_id
func (h *BlackboxHandler) GetUserTrails(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	trails, err := h.postgres.GetBlackboxTrails(c.Request.Context(), userID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get trails"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"trails":  trails,
	})
}
