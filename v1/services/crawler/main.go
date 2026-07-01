// SERP Crawler (crawler-svc)
// Asynq worker consuming SERP_CRAWL jobs.
// Queries SearXNG per keyword, stores results in TimescaleDB,
// deduplicates URLs (hash-based), enqueues CONTENT_EXTRACTION.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	storage "github.com/lastroseo/services/storage"
)

var (
	searxngURL = getEnv("SEARXNG_URL", "http://localhost:8080")
	redisAddr  = getEnv("REDIS_ADDR", "localhost:6379")

	pool     *pgxpool.Pool
	asynqCli *asynq.Client
	asynqSrv *asynq.Server
	visited  sync.Map // URL hash → struct{}
	ticker   = time.NewTicker(2 * time.Second) // ~30 req/min
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
		log.Fatalf("crawler: pg: %v", err)
	}
	defer pool.Close()
	log.Println("   PostgreSQL connected")

	asynqCli = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqCli.Close()

	asynqSrv = asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10, Queues: map[string]int{"default": 3, "low": 1}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("SERP_CRAWL", handleSERPCrawl)

	log.Printf("🕷️  SERP Crawler — SearXNG: %s, workers: 10", searxngURL)

	go func() {
		if err := asynqSrv.Run(mux); err != nil {
			log.Fatalf("crawler: asynq: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("crawler: shutting down")
	asynqSrv.Shutdown()
}

type serpCrawlPayload struct {
	ProjectID string   `json:"project_id"`
	Keywords  []string `json:"keywords"`
}

type searxngResult struct {
	URL     string `json:"url"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type searxngResponse struct {
	Results []searxngResult `json:"results"`
}

func handleSERPCrawl(ctx context.Context, t *asynq.Task) error {
	var p serpCrawlPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("crawler: parse: %w", err)
	}
	if len(p.Keywords) == 0 {
		return fmt.Errorf("crawler: empty keywords")
	}
	log.Printf("crawler: SERP_CRAWL project=%s keywords=%d", p.ProjectID, len(p.Keywords))

	// Ensure keywords exist
	for _, kw := range p.Keywords {
		storage.InsertKeyword(ctx, pool, &storage.Keyword{
			ProjectID: p.ProjectID, Keyword: kw, IsSeed: true, Source: strPtr("serp_crawl"),
		})
	}

	// Lookup keyword IDs
	allKWs, _ := storage.ListKeywords(ctx, pool, p.ProjectID, storage.ListOpts{Limit: 200})
	kwMap := map[string]string{}
	for _, k := range allKWs {
		kwMap[k.Keyword] = k.ID
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, kw := range p.Keywords {
		kwID, ok := kwMap[kw]
		if !ok {
			continue
		}
		wg.Add(1)
		go func(keyword, keywordID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			<-ticker.C // rate limit

			results, err := searchSearXNG(keyword)
			if err != nil {
				log.Printf("crawler: %q: %v", keyword, err)
				return
			}

			now := time.Now()
			stored := 0
			for i, r := range results {
				if i >= 10 {
					break
				}
				if err := storage.InsertSERPResult(ctx, pool, &storage.SERPResult{
					KeywordID: keywordID, Position: i + 1, URL: r.URL,
					Title: r.Title, Snippet: r.Content, CrawledAt: now,
				}); err != nil {
					log.Printf("crawler: insert serp: %v", err)
					continue
				}
				stored++

				// URL dedup
				h := hashURL(r.URL)
				if _, seen := visited.LoadOrStore(h, struct{}{}); seen {
					continue
				}

				// Enqueue content extraction
				payload, _ := json.Marshal(map[string]interface{}{
					"project_id": p.ProjectID, "url": r.URL,
					"keyword": keyword, "position": i + 1,
				})
				asynqCli.Enqueue(asynq.NewTask("CONTENT_EXTRACTION", payload, asynq.Queue("low")))
			}
			log.Printf("crawler: %q → %d serp, %d new urls", keyword, len(results), stored)
		}(kw, kwID)
	}
	wg.Wait()
	log.Printf("crawler: SERP_CRAWL done project=%s", p.ProjectID)

	// Enqueue analytics
	apayload, _ := json.Marshal(map[string]string{"project_id": p.ProjectID})
	_, err := asynqCli.Enqueue(asynq.NewTask("ANALYTICS", apayload, asynq.Queue("default")))
	if err != nil {
		log.Printf("crawler: enqueue ANALYTICS: %v", err)
	} else {
		log.Printf("crawler: enqueued ANALYTICS for project=%s", p.ProjectID)
	}

	return nil
}

func searchSearXNG(keyword string) ([]searxngResult, error) {
	u := fmt.Sprintf("%s/search?q=%s&format=json&categories=general&language=pt-BR",
		searxngURL, url.QueryEscape(keyword))
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get(u)
	if err != nil {
		return nil, fmt.Errorf("searxng: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("searxng: status %d", resp.StatusCode)
	}
	var data searxngResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("searxng: decode: %w", err)
	}
	return data.Results, nil
}

func hashURL(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func strPtr(s string) *string { return &s }
