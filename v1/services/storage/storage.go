// Package storage provides PostgreSQL 16 + TimescaleDB access for all LastroSEO services.
//
// Usage:
//
//	pool, _ := storage.NewPool(ctx, storage.ConfigFromEnv())
//	defer pool.Close()
//	storage.Migrate(ctx, pool)
package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

func ConfigFromEnv() Config {
	return Config{
		Host:     getEnv("POSTGRES_HOST", "localhost"),
		Port:     getEnvInt("POSTGRES_PORT", 5432),
		User:     getEnv("POSTGRES_USER", "lastroseo"),
		Password: getEnv("POSTGRES_PASSWORD", "secret"),
		DBName:   getEnv("POSTGRES_DB", "lastroseo"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.User, c.Password, c.Host, c.Port, c.DBName)
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("storage: parse config: %w", err)
	}
	poolCfg.MaxConns = 20
	poolCfg.MinConns = 2
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("storage: create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("storage: ping: %w", err)
	}
	return pool, nil
}

func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	var n int
	fmt.Sscanf(v, "%d", &n)
	return n
}
