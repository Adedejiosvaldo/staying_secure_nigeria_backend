package handlers

import (
	"net/http"
	"time"

	"github.com/adedejiosvaldo/safetrace/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TripHandler struct {
	tripService *services.TripService
}

func NewTripHandler(ts *services.TripService) *TripHandler {
	return &TripHandler{
		tripService: ts,
	}
}

type StartTripRequest struct {
	UserID           string    `json:"user_id" binding:"required"`
	Destination      string    `json:"destination" binding:"required"`
	EstimatedArrival time.Time `json:"estimated_arrival" binding:"required"`
}

func (h *TripHandler) StartTrip(c *gin.Context) {
	var req StartTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	trip, err := h.tripService.StartTrip(c.Request.Context(), userID, req.Destination, req.EstimatedArrival)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start trip"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trip started successfully",
		"trip":    trip,
	})
}

type LocationRequest struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

func (h *TripHandler) StreamLocation(c *gin.Context) {
	tripIDStr := c.Param("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}

	var req LocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tripService.AddLocation(c.Request.Context(), tripID, req.Lat, req.Lng); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *TripHandler) EndTrip(c *gin.Context) {
	tripIDStr := c.Param("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}

	if err := h.tripService.EndTrip(c.Request.Context(), tripID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to end trip"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Trip ended"})
}

func (h *TripHandler) GetGuardians(c *gin.Context) {
	tripIDStr := c.Param("id")
	tripID, err := uuid.Parse(tripIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trip ID"})
		return
	}

	guardians, err := h.tripService.GetActiveGuardians(c.Request.Context(), tripID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get guardians"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"guardians": guardians})
}
