package handler

import (
	"net/http"
	"strconv"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type ParticipantHandler struct {
	service domain.ParticipantService
}

// NewParticipantHandler creates a new ParticipantHandler instance.
func NewParticipantHandler(service domain.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{
		service: service,
	}
}

// CreateParticipantRequest represents JSON body for creating a participant.
type CreateParticipantRequest struct {
	FullName   string `json:"full_name" binding:"required"`
	CategoryID int    `json:"category_id" binding:"required"`
}

// Create handles POST /api/participants.
func (h *ParticipantHandler) Create(c *gin.Context) {
	var req CreateParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p, err := h.service.Create(c.Request.Context(), req.FullName, req.CategoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

// GetByID handles GET /api/participants/:id.
func (h *ParticipantHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	p, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "participant not found"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// List handles GET /api/participants.
func (h *ParticipantHandler) List(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	list, total, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   list,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GenerateBadge handles GET /api/badges/:participantId/generate.
func (h *ParticipantHandler) GenerateBadge(c *gin.Context) {
	id := c.Param("participantId")
	p, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "participant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"badge_type":       "MOCK_PDF_STUB",
		"participant_id":   p.ID,
		"full_name":        p.FullName,
		"category_id":      p.CategoryID,
		"qr_token":         p.QRToken,
		"download_message": "PDF design is pending implementation. Returning JSON representation.",
	})
}
