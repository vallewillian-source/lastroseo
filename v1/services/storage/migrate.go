package storage

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_init.sql
var initSQL string

//go:embed migrations/002_gaps.sql
var gapsSQL string

//go:embed migrations/003_competitors.sql
var competitorsSQL string

// Migrate runs the embedded schema migration. Idempotent (CREATE TABLE IF NOT EXISTS).
func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, initSQL); err != nil {
		return fmt.Errorf("storage: migrate(001): %w", err)
	}
	if _, err := pool.Exec(ctx, gapsSQL); err != nil {
		return fmt.Errorf("storage: migrate(002): %w", err)
	}
	if _, err := pool.Exec(ctx, competitorsSQL); err != nil {
		return fmt.Errorf("storage: migrate(003): %w", err)
	}
	return nil
}
