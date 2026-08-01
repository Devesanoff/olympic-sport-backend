package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type SyncHandler struct {
	syncService domain.SyncService
}

func NewSyncHandler(syncService domain.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

func (h *SyncHandler) GetOfflinePackage(c *gin.Context) {
	pkg, err := h.syncService.GetOfflinePackage(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate offline package"})
		return
	}

	jsonData, err := json.Marshal(pkg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode offline package"})
		return
	}

	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(jsonData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compress offline package"})
		return
	}
	if err := gz.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to finalize compression"})
		return
	}

	c.Header("Content-Encoding", "gzip")
	c.Data(http.StatusOK, "application/json", b.Bytes())
}

func (h *SyncHandler) UploadLogs(c *gin.Context) {
	var req domain.SyncUploadLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	if err := h.syncService.UploadLogs(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process log upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logs successfully uploaded"})
}
