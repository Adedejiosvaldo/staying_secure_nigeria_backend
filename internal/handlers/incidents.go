package handlers

import (
	"net/http"

	"github.com/adedejiosvaldo/safetrace/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IncidentHandler struct {
	incidentService *services.IncidentService
}

func NewIncidentHandler(is *services.IncidentService) *IncidentHandler {
	return &IncidentHandler{
		incidentService: is,
	}
}

type ReportIncidentRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	HazardType  string  `json:"hazard_type" binding:"required"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat" binding:"required"`
	Lng         float64 `json:"lng" binding:"required"`
}

func (h *IncidentHandler) ReportIncident(c *gin.Context) {
	var req ReportIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	incident, err := h.incidentService.ReportIncident(c.Request.Context(), userID, req.HazardType, req.Description, req.Lat, req.Lng)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to report incident"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Incident reported successfully",
		"incident": incident,
	})
}
