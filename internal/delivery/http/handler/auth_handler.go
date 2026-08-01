package handler

import (
	"net/http"

	"github.com/Devesanoff/olympic-sport-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// AuthHandler holds dependencies for authentication HTTP endpoints.
type AuthHandler struct {
	jwtHelper *jwt.Helper
}

// NewAuthHandler creates a new AuthHandler instance.
func NewAuthHandler(jwtHelper *jwt.Helper) *AuthHandler {
	return &AuthHandler{
		jwtHelper: jwtHelper,
	}
}

// LoginRequest defines parameters required for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse returns the signed access token to the client.
type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

// Login verifies credentials against static accounts and returns a JWT access token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID string
	var role string

	// Static credentials validation for scanning & admin access
	if req.Email == "admin@olympic.com" && req.Password == "admin123" {
		userID = "00000000-0000-0000-0000-000000000001"
		role = "ADMIN"
	} else if req.Email == "scanner@olympic.com" && req.Password == "scanner123" {
		userID = "00000000-0000-0000-0000-000000000002"
		role = "SCANNER"
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	token, err := h.jwtHelper.GenerateToken(userID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken: token,
	})
}
