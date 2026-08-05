package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// RBACMiddlewareConfig configures dependencies for the RBAC middleware.
type RBACMiddlewareConfig struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

// RBACMiddleware checks if the authenticated user's role has the required permission.
// It leverages Redis sets for caching permission checks to maintain low latency.
func RBACMiddleware(cfg *RBACMiddlewareConfig) func(requiredPermission string) gin.HandlerFunc {
	return func(requiredPermission string) gin.HandlerFunc {
		return func(c *gin.Context) {
			roleVal, exists := c.Get(CtxUserRoleKey)
			if !exists {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user role is missing from context"})
				return
			}

			role, ok := roleVal.(string)
			if !ok || role == "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid user role in context"})
				return
			}

			ctx := c.Request.Context()
			hasPermission, err := checkRolePermission(ctx, cfg.Redis, cfg.DB, role, requiredPermission)
			if err != nil {
				log.Error().Err(err).Str("role", role).Str("permission", requiredPermission).Msg("RBAC authorization error")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization lookup failure"})
				return
			}

			if !hasPermission {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}

			c.Next()
		}
	}
}

// checkRolePermission retrieves permissions for a role, checking Redis first, then fallback to Postgres.
func checkRolePermission(ctx context.Context, rdb *redis.Client, db *pgxpool.Pool, role string, requiredPermission string) (bool, error) {
	cacheKey := fmt.Sprintf("role:permissions:%s", role)

	if rdb != nil {
		exists, err := rdb.Exists(ctx, cacheKey).Result()
		if err == nil && exists > 0 {
			hasPerm, err := rdb.SIsMember(ctx, cacheKey, requiredPermission).Result()
			if err == nil {
				return hasPerm, nil
			}
			log.Warn().Err(err).Msg("Redis SISMEMBER failed, falling back to database")
		}
	}

	// Cache Miss / Fallback
	log.Debug().Str("role", role).Msg("Fetching role permissions from PostgreSQL database")

	query := `
		SELECT p.name 
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN roles r ON r.id = rp.role_id
		WHERE r.name = $1;
	`

	var permissions []interface{}
	var hasPerm bool

	if db != nil {
		rows, err := db.Query(ctx, query, role)
		if err != nil {
			return false, fmt.Errorf("database query error: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var permName string
			if err := rows.Scan(&permName); err != nil {
				return false, fmt.Errorf("failed to scan permission row: %w", err)
			}
			permissions = append(permissions, permName)
			if permName == requiredPermission {
				hasPerm = true
			}
		}

		if err = rows.Err(); err != nil {
			return false, fmt.Errorf("row iteration error: %w", err)
		}
	} else {
		// If DB is offline (degraded setup for testing), we mock roles.
		// Keep this map in sync with the seeder's rolePermissions definition.
		log.Warn().Msg("PostgreSQL client not initialized; running mocked RBAC permissions")
		mockedPermissions := map[string][]string{
			"SuperAdmin": {
				"dashboard:view", "scanner:access", "participants:write", "participants:read",
				"reports:read", "zones:read", "zones:write", "categories:read", "categories:write",
				"meal_schedules:read", "meal_schedules:write", "roles:read", "roles:write",
				"users:read", "users:write",
			},
			"ADMIN": {
				"dashboard:view", "scanner:access", "participants:write", "participants:read",
				"reports:read", "zones:read", "zones:write", "categories:read", "categories:write",
				"meal_schedules:read", "meal_schedules:write", "roles:read", "roles:write",
				"users:read", "users:write",
			},
			// Guard: mobile operator — must be able to reach GET /api/sync/offline-package
			"Guard": {
				"scanner:access", "participants:read", "zones:read",
			},
			"SCANNER": {
				"scanner:access", "participants:read", "zones:read",
			},
			"KitchenManager": {
				"scanner:access", "participants:read", "meal_schedules:read", "meal_schedules:write",
			},
		}
		for _, permName := range mockedPermissions[role] {
			permissions = append(permissions, permName)
			if permName == requiredPermission {
				hasPerm = true
			}
		}
	}

	// Save to Redis Cache (SADD with expiration)
	if rdb != nil {
		cachePermissions := permissions
		if len(cachePermissions) == 0 {
			// Sentinel value to prevent cache penetration / DB stampede
			cachePermissions = append(cachePermissions, "__none__")
		}

		pipe := rdb.Pipeline()
		pipe.SAdd(ctx, cacheKey, cachePermissions...)
		pipe.Expire(ctx, cacheKey, 1*time.Hour)
		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to save role permissions cache to Redis")
		}
	}

	return hasPerm, nil
}
