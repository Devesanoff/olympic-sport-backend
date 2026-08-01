package http

import (
	"context"
	"net/http"

	"github.com/Devesanoff/olympic-sport-backend/config"
	"github.com/Devesanoff/olympic-sport-backend/internal/delivery/http/handler"
	"github.com/Devesanoff/olympic-sport-backend/internal/delivery/http/middleware"
	"github.com/Devesanoff/olympic-sport-backend/internal/domain"
	"github.com/Devesanoff/olympic-sport-backend/internal/repository/postgres"
	"github.com/Devesanoff/olympic-sport-backend/internal/service"
	"github.com/Devesanoff/olympic-sport-backend/pkg/hmac"
	"github.com/Devesanoff/olympic-sport-backend/pkg/jwt"
	"github.com/Devesanoff/olympic-sport-backend/pkg/websocket"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// RouterConfig holds dependencies required by HTTP router.
type RouterConfig struct {
	Config     *config.Config
	DB         *pgxpool.Pool
	Redis      *redis.Client
	JWTHelper  *jwt.Helper
	HMACHelper *hmac.Helper
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

	participantRepo := postgres.NewParticipantRepository(cfg.DB)
	participantService := service.NewParticipantService(participantRepo, cfg.HMACHelper)
	participantHandler := handler.NewParticipantHandler(participantService)

	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := handler.NewDashboardWSHandler(wsHub, cfg.JWTHelper)

	scanRepo := postgres.NewScanRepository(cfg.DB)
	scanService := service.NewScanService(scanRepo, scanRepo, cfg.Redis, cfg.HMACHelper, wsHub)
	scanHandler := handler.NewScanHandler(scanService)

	syncRepo := postgres.NewSyncRepository(cfg.DB)
	syncService := service.NewSyncService(syncRepo)
	syncHandler := handler.NewSyncHandler(syncService)

	badgeService := service.NewBadgeService(participantRepo, scanRepo)
	badgeHandler := handler.NewBadgeHandler(badgeService)

	adminRepo := postgres.NewAdminRepository(cfg.DB)
	adminService := service.NewAdminService(&domain.AdminRepoBundle{
		ZoneRepo:         adminRepo,
		CategoryRepo:     adminRepo,
		MealScheduleRepo: adminRepo,
		RBACRepo:         adminRepo,
		UserRepo:         adminRepo,
	}, cfg.Redis)
	adminHandler := handler.NewAdminHandler(
		adminService,
		adminService,
		adminService,
		adminService,
		adminService,
	)

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

		// WebSocket Live Stats (using standard route group but handles WS upgrade)
		protected.GET("/dashboard/live-stats", wsHandler.ServeLiveStatsWS)

		protected.POST("/scanner/scan", rbac("scanner:access"), func(c *gin.Context) {
			userID, _ := c.Get(middleware.CtxUserIDKey)
			c.JSON(http.StatusOK, gin.H{
				"message": "Scan verification mock endpoint",
				"user_id": userID,
			})
		})

		// High-Throughput Scan Endpoints
		protected.POST("/scan/access", rbac("scanner:access"), scanHandler.ScanAccess)
		protected.POST("/scan/meal", rbac("scanner:access"), scanHandler.ScanMeal)

		// Sync Endpoints
		protected.GET("/sync/offline-package", rbac("scanner:access"), syncHandler.GetOfflinePackage)
		protected.POST("/sync/upload-logs", rbac("scanner:access"), syncHandler.UploadLogs)

		// Participant Endpoints
		protected.POST("/participants", rbac("participants:write"), participantHandler.Create)
		protected.GET("/participants/:id", rbac("participants:read"), participantHandler.GetByID)
		protected.GET("/participants", rbac("participants:read"), participantHandler.List)
		
		// Badge Endpoints
		protected.GET("/badges/:participantId/generate", rbac("participants:read"), badgeHandler.GenerateSingle)
		protected.POST("/badges/bulk-generate", rbac("participants:read"), badgeHandler.GenerateBulk)

		// Admin CRUD Endpoints: Zones
		protected.GET("/zones", rbac("zones:read"), adminHandler.ListZones)
		protected.GET("/zones/:id", rbac("zones:read"), adminHandler.GetZoneByID)
		protected.POST("/zones", rbac("zones:write"), adminHandler.CreateZone)
		protected.PUT("/zones/:id", rbac("zones:write"), adminHandler.UpdateZone)
		protected.DELETE("/zones/:id", rbac("zones:write"), adminHandler.DeleteZone)

		// Admin CRUD Endpoints: Categories
		protected.GET("/categories", rbac("categories:read"), adminHandler.ListCategories)
		protected.GET("/categories/:id", rbac("categories:read"), adminHandler.GetCategoryByID)
		protected.POST("/categories", rbac("categories:write"), adminHandler.CreateCategory)
		protected.PUT("/categories/:id", rbac("categories:write"), adminHandler.UpdateCategory)
		protected.DELETE("/categories/:id", rbac("categories:write"), adminHandler.DeleteCategory)
		protected.POST("/categories/:id/zones", rbac("categories:write"), adminHandler.SetCategoryAllowedZones)

		// Admin CRUD Endpoints: Meal Schedules
		protected.GET("/meal-schedules", rbac("meal_schedules:read"), adminHandler.ListMealSchedules)
		protected.GET("/meal-schedules/:id", rbac("meal_schedules:read"), adminHandler.GetMealScheduleByID)
		protected.POST("/meal-schedules", rbac("meal_schedules:write"), adminHandler.CreateMealSchedule)
		protected.PUT("/meal-schedules/:id", rbac("meal_schedules:write"), adminHandler.UpdateMealSchedule)
		protected.DELETE("/meal-schedules/:id", rbac("meal_schedules:write"), adminHandler.DeleteMealSchedule)
		protected.POST("/meal-schedules/:id/categories", rbac("meal_schedules:write"), adminHandler.SetMealScheduleCategories)

		// Admin CRUD Endpoints: Dynamic RBAC
		protected.GET("/roles", rbac("roles:read"), adminHandler.ListRoles)
		protected.POST("/roles", rbac("roles:write"), adminHandler.CreateRole)
		protected.DELETE("/roles/:id", rbac("roles:write"), adminHandler.DeleteRole)
		protected.GET("/permissions", rbac("roles:read"), adminHandler.ListPermissions)
		protected.POST("/permissions", rbac("roles:write"), adminHandler.CreatePermission)
		protected.POST("/roles/:id/permissions", rbac("roles:write"), adminHandler.AssignPermissionsToRole)

		// Admin CRUD Endpoints: Users Management
		protected.GET("/users", rbac("users:read"), adminHandler.ListUsers)
		protected.GET("/users/:id", rbac("users:read"), adminHandler.GetUserByID)
		protected.POST("/users", rbac("users:write"), adminHandler.CreateUser)
		protected.PUT("/users/:id", rbac("users:write"), adminHandler.UpdateUser)
		protected.DELETE("/users/:id", rbac("users:write"), adminHandler.DeleteUser)
		protected.POST("/users/:id/roles", rbac("users:write"), adminHandler.AssignUserRoles)
	}

	return router
}
