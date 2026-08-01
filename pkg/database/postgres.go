package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Devesanoff/olympic-sport-backend/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// NewPostgresPool initializes and returns a high-concurrency pgxpool PostgreSQL connection pool.
func NewPostgresPool(ctx context.Context, cfg *config.PostgresConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres DSN: %w", err)
	}

	// Pool tuning for high-concurrency scanner workloads
	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(pingCtx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create postgres pool: %w", err)
	}

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}

	log.Info().
		Int32("max_conns", poolConfig.MaxConns).
		Int32("min_conns", poolConfig.MinConns).
		Msg("Successfully connected to PostgreSQL database")

	return pool, nil
}
