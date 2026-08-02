package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	service domain.ReportService
}

// NewReportHandler creates a new ReportHandler.
func NewReportHandler(service domain.ReportService) *ReportHandler {
	return &ReportHandler{
		service: service,
	}
}

// GetAccessLogs handles GET /api/reports/access-logs.
func (h *ReportHandler) GetAccessLogs(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	zoneIDStr := c.Query("zone_id")
	categoryIDStr := c.Query("category_id")
	statusStr := c.Query("status")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	startDate, err := parseTime(startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD or RFC3339"})
		return
	}

	endDate, err := parseTime(endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD or RFC3339"})
		return
	}

	var zoneID *int
	if zoneIDStr != "" {
		zid, err := strconv.Atoi(zoneIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone_id"})
			return
		}
		zoneID = &zid
	}

	var categoryID *int
	if categoryIDStr != "" {
		cid, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		categoryID = &cid
	}

	var status *string
	if statusStr != "" {
		statusStrUpper := strings.ToUpper(statusStr)
		if statusStrUpper != "ALLOWED" && statusStrUpper != "DENIED" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be ALLOWED or DENIED"})
			return
		}
		status = &statusStrUpper
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	filter := &domain.AccessLogFilter{
		StartDate:  startDate,
		EndDate:    endDate,
		ZoneID:     zoneID,
		CategoryID: categoryID,
		Status:     status,
		Limit:      limit,
		Offset:     offset,
	}

	list, total, err := h.service.GetAccessLogs(c.Request.Context(), filter)
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

// GetMealLogs handles GET /api/reports/meal-logs.
func (h *ReportHandler) GetMealLogs(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	categoryIDStr := c.Query("category_id")
	statusStr := c.Query("status")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	startDate, err := parseTime(startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD or RFC3339"})
		return
	}

	endDate, err := parseTime(endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, use YYYY-MM-DD or RFC3339"})
		return
	}

	var categoryID *int
	if categoryIDStr != "" {
		cid, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		categoryID = &cid
	}

	var status *string
	if statusStr != "" {
		statusStrUpper := strings.ToUpper(statusStr)
		if statusStrUpper != "ALLOWED" && statusStrUpper != "DENIED" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status must be ALLOWED or DENIED"})
			return
		}
		status = &statusStrUpper
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	filter := &domain.MealLogFilter{
		StartDate:  startDate,
		EndDate:    endDate,
		CategoryID: categoryID,
		Status:     status,
		Limit:      limit,
		Offset:     offset,
	}

	list, total, err := h.service.GetMealLogs(c.Request.Context(), filter)
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

// GetDeniedAttempts handles GET /api/reports/denied-attempts.
func (h *ReportHandler) GetDeniedAttempts(c *gin.Context) {
	list, err := h.service.GetDeniedAttempts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

// ExportExcel handles GET /api/reports/export/excel.
func (h *ReportHandler) ExportExcel(c *gin.Context) {
	reportType := c.Query("type")
	if reportType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing type parameter"})
		return
	}

	var filter interface{}

	switch strings.ToLower(reportType) {
	case "access":
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")
		zoneIDStr := c.Query("zone_id")
		categoryIDStr := c.Query("category_id")
		statusStr := c.Query("status")

		startDate, err := parseTime(startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
			return
		}

		endDate, err := parseTime(endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
			return
		}

		var zoneID *int
		if zoneIDStr != "" {
			zid, err := strconv.Atoi(zoneIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone_id"})
				return
			}
			zoneID = &zid
		}

		var categoryID *int
		if categoryIDStr != "" {
			cid, err := strconv.Atoi(categoryIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
				return
			}
			categoryID = &cid
		}

		var status *string
		if statusStr != "" {
			statusStrUpper := strings.ToUpper(statusStr)
			if statusStrUpper != "ALLOWED" && statusStrUpper != "DENIED" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "status must be ALLOWED or DENIED"})
				return
			}
			status = &statusStrUpper
		}

		filter = &domain.AccessLogFilter{
			StartDate:  startDate,
			EndDate:    endDate,
			ZoneID:     zoneID,
			CategoryID: categoryID,
			Status:     status,
		}

	case "meal":
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")
		categoryIDStr := c.Query("category_id")
		statusStr := c.Query("status")

		startDate, err := parseTime(startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
			return
		}

		endDate, err := parseTime(endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
			return
		}

		var categoryID *int
		if categoryIDStr != "" {
			cid, err := strconv.Atoi(categoryIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
				return
			}
			categoryID = &cid
		}

		var status *string
		if statusStr != "" {
			statusStrUpper := strings.ToUpper(statusStr)
			if statusStrUpper != "ALLOWED" && statusStrUpper != "DENIED" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "status must be ALLOWED or DENIED"})
				return
			}
			status = &statusStrUpper
		}

		filter = &domain.MealLogFilter{
			StartDate:  startDate,
			EndDate:    endDate,
			CategoryID: categoryID,
			Status:     status,
		}

	case "denied":
		// No filters needed

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report type"})
		return
	}

	data, err := h.service.ExportExcel(c.Request.Context(), reportType, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := fmt.Sprintf("%s_report_%s.xlsx", reportType, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// parseTime parses YYYY-MM-DD, RFC3339, or YYYY-MM-DD HH:MM:SS format strings.
func parseTime(val string) (*time.Time, error) {
	if val == "" {
		return nil, nil
	}

	// Try RFC3339
	t, err := time.Parse(time.RFC3339, val)
	if err == nil {
		return &t, nil
	}

	// Try YYYY-MM-DD
	t, err = time.Parse("2006-01-02", val)
	if err == nil {
		return &t, nil
	}

	// Try YYYY-MM-DD HH:MM:SS
	t, err = time.Parse("2006-01-02 15:04:05", val)
	if err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("invalid time format")
}
