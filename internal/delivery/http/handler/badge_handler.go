package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type BadgeHandler struct {
	badgeService domain.BadgeService
}

func NewBadgeHandler(badgeService domain.BadgeService) *BadgeHandler {
	return &BadgeHandler{
		badgeService: badgeService,
	}
}

// GenerateSingle handles GET /api/badges/:participantId/generate
func (h *BadgeHandler) GenerateSingle(c *gin.Context) {
	participantID := c.Param("participantId")

	pdfBytes, err := h.badgeService.GenerateSingleBadge(c.Request.Context(), participantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("badge_%s.pdf", participantID)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GenerateBulk handles POST /api/badges/bulk-generate
func (h *BadgeHandler) GenerateBulk(c *gin.Context) {
	var req domain.BulkBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	pdfBytes, err := h.badgeService.GenerateBulkBadges(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("badges_bulk_%d.pdf", time.Now().Unix())
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
