package handlers

import (
	"net/http"

	"github.com/adedejiosvaldo/safetrace/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TimerHandler struct {
	timerService *services.TimerService
}

func NewTimerHandler(ts *services.TimerService) *TimerHandler {
	return &TimerHandler{
		timerService: ts,
	}
}

type StartTimerRequest struct {
	UserID          string  `json:"user_id" binding:"required"`
	DurationSeconds int     `json:"duration_seconds" binding:"required"`
	Label           string  `json:"label" binding:"required"`
	Lat             float64 `json:"lat" binding:"required"`
	Lng             float64 `json:"lng" binding:"required"`
}

func (h *TimerHandler) StartTimer(c *gin.Context) {
	var req StartTimerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	timer, err := h.timerService.StartTimer(c.Request.Context(), userID, req.DurationSeconds, req.Label, req.Lat, req.Lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start timer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Timer started successfully",
		"timer":   timer,
	})
}

type ExtendTimerRequest struct {
	AdditionalSeconds int `json:"additional_seconds" binding:"required"`
}

func (h *TimerHandler) ExtendTimer(c *gin.Context) {
	timerIDStr := c.Param("id")
	timerID, err := uuid.Parse(timerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timer ID"})
		return
	}

	var req ExtendTimerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.timerService.ExtendTimer(c.Request.Context(), timerID, req.AdditionalSeconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extend timer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Timer extended"})
}

func (h *TimerHandler) MarkSafe(c *gin.Context) {
	timerIDStr := c.Param("id")
	timerID, err := uuid.Parse(timerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timer ID"})
		return
	}

	if err := h.timerService.MarkSafe(c.Request.Context(), timerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark safe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Timer cancelled, marked safe"})
}
