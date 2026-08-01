package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Devesanoff/olympic-sport-backend/config"
	httpDelivery "github.com/Devesanoff/olympic-sport-backend/internal/delivery/http"
	"github.com/Devesanoff/olympic-sport-backend/pkg/database"
	"github.com/Devesanoff/olympic-sport-backend/pkg/hmac"
	"github.com/Devesanoff/olympic-sport-backend/pkg/jwt"
	"github.com/Devesanoff/olympic-sport-backend/pkg/logger"
	"github.com/rs/zerolog/log"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load environment configuration")
	}

	// 2. Initialize Structured Logger
	logger.InitLogger(cfg.App.LogLevel, cfg.App.Env)
	log.Info().
		Str("env", cfg.App.Env).
		Str("log_level", cfg.App.LogLevel).
		Msg("Starting Accreditation and Access/Meal Control System API")

	ctx := context.Background()

	// 3. Connect PostgreSQL Database Pool
	dbPool, err := database.NewPostgresPool(ctx, &cfg.Postgres)
	if err != nil {
		log.Warn().Err(err).Msg("PostgreSQL connection failed (system will run in degraded mode if DB offline)")
	} else {
		defer dbPool.Close()
	}

	// 4. Connect Redis Client
	redisClient, err := database.NewRedisClient(ctx, &cfg.Redis)
	if err != nil {
		log.Warn().Err(err).Msg("Redis connection failed (system will run in degraded mode if Redis offline)")
	} else {
		defer func() {
			if closeErr := redisClient.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("Error closing Redis client")
			}
		}()
	}

	// 5. Initialize JWT and HMAC Helpers
	jwtHelper := jwt.NewHelper(cfg.JWT.Secret, cfg.JWT.Expiration)
	hmacHelper := hmac.NewHelper(cfg.HMAC.Secret)

	// 6. Initialize Router & Server
	router := httpDelivery.NewRouter(&httpDelivery.RouterConfig{
		Config:     cfg,
		DB:         dbPool,
		Redis:      redisClient,
		JWTHelper:  jwtHelper,
		HMACHelper: hmacHelper,
	})

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// 6. Start HTTP Server asynchronously
	go func() {
		log.Info().Str("port", cfg.Server.Port).Msg("HTTP Server is listening")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("HTTP Server encountered an unexpected error")
		}
	}()

	// 7. Listen for OS Graceful Shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("Shutdown signal received. Initiating graceful shutdown...")

	// 8. Graceful Shutdown Timeout Context
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP Server forced to shutdown with error")
	} else {
		log.Info().Msg("HTTP Server stopped gracefully")
	}

	log.Info().Msg("Application shutdown complete")
}
