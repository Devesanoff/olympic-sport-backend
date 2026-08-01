package handler

import (
	"net/http"

	"github.com/Devesanoff/olympic-sport-backend/pkg/jwt"
	"github.com/Devesanoff/olympic-sport-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
	gorillaWs "github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for the dashboard, restrict in production if necessary.
		return true
	},
}

type DashboardWSHandler struct {
	hub       *websocket.Hub
	jwtHelper *jwt.Helper
}

func NewDashboardWSHandler(hub *websocket.Hub, jwtHelper *jwt.Helper) *DashboardWSHandler {
	return &DashboardWSHandler{
		hub:       hub,
		jwtHelper: jwtHelper,
	}
}

func (h *DashboardWSHandler) ServeLiveStatsWS(c *gin.Context) {
	// 1. Authenticate via token query parameter (standard for JS WebSockets)
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token parameter"})
		return
	}

	claims, err := h.jwtHelper.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	// Optional: Check if the user has admin/dashboard permissions based on claims here
	log.Info().Str("user_id", claims.UserID).Msg("WebSocket connection authenticated")

	// 2. Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return
	}

	// 3. Register Client to Hub
	websocket.RegisterClient(h.hub, conn)
}
