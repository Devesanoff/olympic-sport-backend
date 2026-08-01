package middleware

import (
	"net/http"
	"strings"

	"github.com/Devesanoff/olympic-sport-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey = "authorization"
	AuthorizationTypeBearer = "bearer"
	CtxUserIDKey           = "user_id"
	CtxUserRoleKey         = "role"
)

// AuthMiddleware validates JWT Bearer tokens and injects claim context.
func AuthMiddleware(jwtHelper *jwt.Helper) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is missing"})
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unsupported authorization type"})
			return
		}

		accessToken := fields[1]
		claims, err := jwtHelper.ValidateToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUserRoleKey, claims.Role)
		c.Next()
	}
}
