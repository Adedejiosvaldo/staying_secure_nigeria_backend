package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/adedejiosvaldo/safetrace/backend/internal/config"
	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
)

type ContactsHandler struct {
	cfg      *config.Config
	postgres *database.PostgresDB
}

func NewContactsHandler(
	cfg *config.Config,
	postgres *database.PostgresDB,
) *ContactsHandler {
	return &ContactsHandler{
		cfg:      cfg,
		postgres: postgres,
	}
}

type AddContactRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type UpdateContactRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

// GET /v1/user/:id/contacts
func (h *ContactsHandler) GetContacts(c *gin.Context) {
	userIDStr := c.Param("id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid user_id '%s': %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id",
			"details": err.Error(),
		})
		return
	}

	// Verify user exists
	user, err := h.postgres.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		log.Printf("ERROR: Failed to get user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
			"details": err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Return contacts from user's trusted_contacts JSON field
	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"contacts": user.TrustedContacts,
	})
}

// POST /v1/user/:id/contacts
func (h *ContactsHandler) AddContact(c *gin.Context) {
	userIDStr := c.Param("id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid user_id '%s': %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id",
			"details": err.Error(),
		})
		return
	}

	var req AddContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	log.Printf("INFO: Adding contact for user %s: name=%s, phone=%s", userID, req.Name, req.Phone)

	contact := map[string]string{
		"id":    uuid.New().String(),
		"name":  req.Name,
		"phone": req.Phone,
	}

	if err := h.postgres.AddContact(c.Request.Context(), userID, contact); err != nil {
		log.Printf("ERROR: Failed to add contact: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to add contact",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"contact": contact,
		"message": "contact added successfully",
	})
}

// PUT /v1/user/:id/contacts/:contactId
func (h *ContactsHandler) UpdateContact(c *gin.Context) {
	userIDStr := c.Param("id")
	contactID := c.Param("contactId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid user_id '%s': %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id",
			"details": err.Error(),
		})
		return
	}

	var req UpdateContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERROR: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
			"details": err.Error(),
		})
		return
	}

	log.Printf("INFO: Updating contact %s for user %s", contactID, userID)

	updates := map[string]string{
		"id": contactID,
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}

	if err := h.postgres.UpdateContact(c.Request.Context(), userID, contactID, updates); err != nil {
		log.Printf("ERROR: Failed to update contact: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update contact",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "contact updated successfully",
	})
}

// DELETE /v1/user/:id/contacts/:contactId
func (h *ContactsHandler) DeleteContact(c *gin.Context) {
	userIDStr := c.Param("id")
	contactID := c.Param("contactId")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("ERROR: Invalid user_id '%s': %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id",
			"details": err.Error(),
		})
		return
	}

	log.Printf("INFO: Deleting contact %s for user %s", contactID, userID)

	if err := h.postgres.DeleteContact(c.Request.Context(), userID, contactID); err != nil {
		log.Printf("ERROR: Failed to delete contact: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete contact",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"message": "contact deleted successfully",
	})
}
