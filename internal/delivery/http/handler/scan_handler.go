package handler

import (
	"net/http"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type ScanHandler struct {
	service domain.ScanService
}

// NewScanHandler creates a new ScanHandler.
func NewScanHandler(service domain.ScanService) *ScanHandler {
	return &ScanHandler{
		service: service,
	}
}

// ScanAccess handles POST /api/scan/access.
func (h *ScanHandler) ScanAccess(c *gin.Context) {
	var req domain.AccessScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.service.ScanAccess(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// ScanMeal handles POST /api/scan/meal.
func (h *ScanHandler) ScanMeal(c *gin.Context) {
	var req domain.MealScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.service.ScanMeal(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
