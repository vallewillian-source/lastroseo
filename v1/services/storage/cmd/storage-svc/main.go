// Storage Service — standalone health server + migration runner.
// Other services import the storage package directly.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	storage "github.com/lastroseo/services/storage"
)

var (
	pgHost    = getEnv("POSTGRES_HOST", "localhost")
	pgPort    = getEnvInt("POSTGRES_PORT", 5432)
	pgUser    = getEnv("POSTGRES_USER", "lastroseo")
	pgPass    = getEnv("POSTGRES_PASSWORD", "secret")
	pgDB      = getEnv("POSTGRES_DB", "lastroseo")
)

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

func main() {
	log.Println("🗄️  Storage Service (storage-svc) starting")

	cfg := storage.Config{
		Host: pgHost, Port: pgPort, User: pgUser, Password: pgPass, DBName: pgDB,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := storage.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("storage: connect: %v", err)
	}
	defer pool.Close()
	log.Printf("   PostgreSQL connected: %s:%d/%s", pgHost, pgPort, pgDB)

	// Run migrations
	if err := storage.Migrate(ctx, pool); err != nil {
		log.Fatalf("storage: migrate: %v", err)
	}
	log.Println("   Migrations applied")

	// Health server
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	http.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		status := "ready"
		if err := storage.HealthCheck(ctx, pool); err != nil {
			status = "not ready"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": status,
		})
	})

	addr := ":8080"
	log.Printf("🚀 Storage service listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("storage: server: %v", err)
	}
}
