package handlers

import (
	"net/http"

	"github.com/adedejiosvaldo/safetrace/backend/internal/models"
	"github.com/adedejiosvaldo/safetrace/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SettingsHandler struct {
	settingsService *services.SettingsService
}

func NewSettingsHandler(ss *services.SettingsService) *SettingsHandler {
	return &SettingsHandler{
		settingsService: ss,
	}
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var newSettings models.UserSettings
	if err := c.ShouldBindJSON(&newSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.settingsService.UpdateUserSettings(c.Request.Context(), userID, newSettings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

func (h *SettingsHandler) DeleteData(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.settingsService.PurgeUserData(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to purge user data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User data purged"})
}
