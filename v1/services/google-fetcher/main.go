// Google Data Fetcher (google-svc)
// Asynq worker consuming KEYWORD_RESEARCH jobs.
// Scrapes Google Autocomplete for keyword expansion (free, no API key),
// stores discovered keywords in PG, enqueues SERP_CRAWL for all keywords.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	storage "github.com/lastroseo/services/storage"
)

var (
	redisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	pool      *pgxpool.Pool
	asynqCli  *asynq.Client
	asynqSrv  *asynq.Server
)

func getEnv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}

func main() {
	cfg := storage.ConfigFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	pool, err = storage.NewPool(ctx, cfg)
	if err != nil {
		log.Fatalf("google-svc: pg: %v", err)
	}
	defer pool.Close()
	log.Println("   PostgreSQL connected")

	asynqCli = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqCli.Close()

	asynqSrv = asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 5, Queues: map[string]int{"default": 3, "low": 1}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("KEYWORD_RESEARCH", handleKeywordResearch)

	log.Println("🔍 Google Fetcher — autocomplete scraper, workers: 5")

	go func() {
		if err := asynqSrv.Run(mux); err != nil {
			log.Fatalf("google-svc: asynq: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("google-svc: shutting down")
	asynqSrv.Shutdown()
}

type keywordResearchPayload struct {
	ProjectID    string   `json:"project_id"`
	Name         string   `json:"name"`
	BusinessDesc string   `json:"business_desc"`
	SeedKeywords []string `json:"seed_keywords"`
}

func handleKeywordResearch(ctx context.Context, t *asynq.Task) error {
	var p keywordResearchPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("google-svc: parse: %w", err)
	}
	if len(p.SeedKeywords) == 0 {
		return fmt.Errorf("google-svc: no seed keywords")
	}
	log.Printf("google-svc: KEYWORD_RESEARCH project=%s seeds=%d", p.ProjectID, len(p.SeedKeywords))

	// Update job
	updateJobStatus(ctx, p.ProjectID, storage.JobProcessing)

	// Insert seeds
	for _, kw := range p.SeedKeywords {
		storage.InsertKeyword(ctx, pool, &storage.Keyword{
			ProjectID: p.ProjectID, Keyword: kw, IsSeed: true, Source: strPtr("seed"),
		})
	}

	// Expand via Google Autocomplete
	allKWs := map[string]bool{}
	for _, kw := range p.SeedKeywords {
		allKWs[kw] = true
	}
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 3)

	for _, seed := range p.SeedKeywords {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			suggestions, err := fetchAutocomplete(s)
			if err != nil {
				log.Printf("google-svc: autocomplete %q: %v", s, err)
				return
			}
			mu.Lock()
			for _, sug := range suggestions {
				if kw := strings.TrimSpace(sug); len(kw) > 2 {
					allKWs[kw] = true
				}
			}
			mu.Unlock()
			time.Sleep(200 * time.Millisecond)
		}(seed)
	}
	wg.Wait()

	// Store all keywords
	allList := make([]string, 0, len(allKWs))
	for kw := range allKWs {
		allList = append(allList, kw)
		storage.InsertKeyword(ctx, pool, &storage.Keyword{
			ProjectID: p.ProjectID, Keyword: kw, IsSeed: false, Source: strPtr("autocomplete"),
		})
	}
	newCount := len(allList) - len(p.SeedKeywords)
	log.Printf("google-svc: %d seeds -> %d total (%d new)", len(p.SeedKeywords), len(allList), newCount)

	// Enqueue SERP_CRAWL
	crawlPayload, _ := json.Marshal(map[string]interface{}{
		"project_id": p.ProjectID, "keywords": allList,
	})
	crawlTask, _ := asynqCli.Enqueue(
		asynq.NewTask("SERP_CRAWL", crawlPayload, asynq.Queue("default")),
	)
	log.Printf("google-svc: SERP_CRAWL %s (%d keywords)", crawlTask.ID, len(allList))

	// Complete
	result, _ := json.Marshal(map[string]int{"total": len(allList), "new": newCount})
	updateJobStatus(ctx, p.ProjectID, storage.JobCompleted)
	_ = result
	return nil
}

func updateJobStatus(ctx context.Context, projectID string, status storage.JobStatus) {
	jobs, _ := storage.ListJobsByProject(ctx, pool, projectID, storage.ListOpts{Limit: 5})
	for _, j := range jobs {
		if j.Type == "KEYWORD_RESEARCH" && j.Status != string(storage.JobCompleted) {
			storage.UpdateJobStatus(ctx, pool, j.ID, status, nil)
			return
		}
	}
}

func fetchAutocomplete(seed string) ([]string, error) {
	u := "https://suggestqueries.google.com/complete/search?client=firefox&q=" + url.QueryEscape(seed)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw) < 2 {
		return nil, nil
	}
	arr, _ := raw[1].([]interface{})
	var out []string
	for _, v := range arr {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out, nil
}

func strPtr(s string) *string { return &s }
