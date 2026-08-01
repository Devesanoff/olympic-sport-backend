package http

import (
	"context"
	"net/http"

	"github.com/Devesanoff/olympic-sport-backend/config"
	"github.com/Devesanoff/olympic-sport-backend/internal/delivery/http/handler"
	"github.com/Devesanoff/olympic-sport-backend/internal/delivery/http/middleware"
	"github.com/Devesanoff/olympic-sport-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// RouterConfig holds dependencies required by HTTP router.
type RouterConfig struct {
	Config    *config.Config
	DB        *pgxpool.Pool
	Redis     *redis.Client
	JWTHelper *jwt.Helper
}

// NewRouter initializes Gin HTTP engine with middleware and core system routes.
func NewRouter(cfg *RouterConfig) *gin.Engine {
	if cfg.Config.Server.Mode != "" {
		gin.SetMode(cfg.Config.Server.Mode)
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())

	// System Health & Liveness Endpoints
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "UP",
			"app":    "olympic-sport-backend",
		})
	})

	router.GET("/readyz", func(c *gin.Context) {
		ctx := context.Background()

		dbStatus := "UP"
		if cfg.DB != nil {
			if err := cfg.DB.Ping(ctx); err != nil {
				dbStatus = "DOWN"
			}
		} else {
			dbStatus = "N/A"
		}

		redisStatus := "UP"
		if cfg.Redis != nil {
			if err := cfg.Redis.Ping(ctx).Err(); err != nil {
				redisStatus = "DOWN"
			}
		} else {
			redisStatus = "N/A"
		}

		status := http.StatusOK
		if dbStatus == "DOWN" || redisStatus == "DOWN" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":   "READY",
			"postgres": dbStatus,
			"redis":    redisStatus,
		})
	})

	// Handlers
	authHandler := handler.NewAuthHandler(cfg.JWTHelper)

	// Public API Group
	api := router.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)
	}

	// RBAC middleware builder
	rbac := middleware.RBACMiddleware(&middleware.RBACMiddlewareConfig{
		DB:    cfg.DB,
		Redis: cfg.Redis,
	})

	// Protected API Group
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware(cfg.JWTHelper))
	{
		protected.GET("/admin/dashboard", rbac("dashboard:view"), func(c *gin.Context) {
			userID, _ := c.Get(middleware.CtxUserIDKey)
			c.JSON(http.StatusOK, gin.H{
				"message": "Welcome to the Admin Dashboard",
				"user_id": userID,
			})
		})

		protected.POST("/scanner/scan", rbac("scanner:access"), func(c *gin.Context) {
			userID, _ := c.Get(middleware.CtxUserIDKey)
			c.JSON(http.StatusOK, gin.H{
				"message": "Scan verification mock endpoint",
				"user_id": userID,
			})
		})
	}

	return router
}
