package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a connection pool for the given DSN. It applies
// conservative pool-lifetime defaults so connections get recycled
// periodically (helpful behind load balancers/poolers that silently drop
// long-lived connections).
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	if cfg.MaxConnLifetime == 0 {
		cfg.MaxConnLifetime = time.Hour
	}
	if cfg.MaxConnIdleTime == 0 {
		cfg.MaxConnIdleTime = 30 * time.Minute
	}
	if cfg.HealthCheckPeriod == 0 {
		cfg.HealthCheckPeriod = time.Minute
	}

	return pgxpool.NewWithConfig(ctx, cfg)
}
