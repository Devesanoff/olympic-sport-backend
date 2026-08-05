package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler holds dependencies for authentication HTTP endpoints.
type AuthHandler struct {
	jwtHelper *jwt.Helper
	userRepo  domain.UserAdminRepository
}

// NewAuthHandler creates a new AuthHandler instance.
// userRepo is used to validate credentials against the database.
func NewAuthHandler(jwtHelper *jwt.Helper, userRepo domain.UserAdminRepository) *AuthHandler {
	return &AuthHandler{
		jwtHelper: jwtHelper,
		userRepo:  userRepo,
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

// Login fetches the user from the DB, verifies the bcrypt password, reads
// their primary role, and issues a signed JWT containing the real role name.
// This ensures that any role (Guard, KitchenManager, SuperAdmin, …) stored in
// the database will be correctly embedded into the token and subsequently
// honoured by RBACMiddleware.
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// --- 1. Fetch user from database ---
	user, err := h.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if isNotFoundError(ctx, err) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify credentials"})
		return
	}

	// --- 2. Verify bcrypt password ---
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		return
	}

	// --- 3. Extract primary role name ---
	// A user may have multiple roles; we embed the first (highest-priority) one.
	// RBACMiddleware looks up all permissions for that role name from the DB/Redis.
	roleName := extractPrimaryRole(user)
	if roleName == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "user has no assigned roles"})
		return
	}

	// --- 4. Issue JWT ---
	token, err := h.jwtHelper.GenerateToken(user.ID, roleName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate authentication token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{AccessToken: token})
}

// extractPrimaryRole returns the name of the first role assigned to the user.
// Returns an empty string if the user has no roles.
func extractPrimaryRole(user *domain.User) string {
	if len(user.Roles) > 0 {
		return user.Roles[0].Name
	}
	return ""
}

// isNotFoundError returns true when the DB signals that no row was found.
func isNotFoundError(_ context.Context, err error) bool {
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	return strings.Contains(err.Error(), "no rows")
}
