package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// NewRedisClient initializes and returns a go-redis client.
func NewRedisClient(ctx context.Context, cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	log.Info().
		Str("addr", cfg.Addr()).
		Int("pool_size", cfg.PoolSize).
		Msg("Successfully connected to Redis")

	return client, nil
}
