// Content Extractor (extractor-svc)
// Asynq worker consuming CONTENT_EXTRACTION jobs.
// Downloads HTML, parses metadata, extracts topics via Gemma LLM,
// stores page_data + keyword_pages.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	llm "github.com/lastroseo/pkg/llm"
	storage "github.com/lastroseo/services/storage"
)

var (
	redisAddr = getEnv("REDIS_ADDR", "localhost:6379")
	pool      *pgxpool.Pool
	asynqCli  *asynq.Client
	asynqSrv  *asynq.Server
	llmClient *llm.Client
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
		log.Fatalf("extractor: pg: %v", err)
	}
	defer pool.Close()
	log.Println("   PostgreSQL connected")

	llmClient = llm.NewClient(os.Getenv("OLLAMA_URL"), "gemma4:e4b")
	log.Printf("   LLM: gemma4:e4b @ %s", os.Getenv("OLLAMA_URL"))

	asynqCli = asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	defer asynqCli.Close()

	asynqSrv = asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10, Queues: map[string]int{"low": 1, "default": 3}},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc("CONTENT_EXTRACTION", handleContentExtraction)

	log.Println("📄 Content Extractor — workers: 10")

	go func() {
		if err := asynqSrv.Run(mux); err != nil {
			log.Fatalf("extractor: asynq: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("extractor: shutting down")
	asynqSrv.Shutdown()
}

type extractionPayload struct {
	ProjectID string `json:"project_id"`
	URL       string `json:"url"`
	Keyword   string `json:"keyword"`
	Position  int    `json:"position"`
}

var projectCount sync.Map

func handleContentExtraction(ctx context.Context, t *asynq.Task) error {
	var p extractionPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("extractor: parse: %w", err)
	}

	html, err := download(p.URL)
	if err != nil {
		log.Printf("extractor: download %s: %v", p.URL, err)
		return nil
	}

	title := extractTitle(html)
	metaDesc := extractMeta(html)
	h1 := extractHeadings(html, "h1")
	h2 := extractHeadings(html, "h2")
	h3 := extractHeadings(html, "h3")
	wordCount := countWords(stripTags(html))
	imgCount := countMatches(html, `<img\s`)
	vidCount := countMatches(html, `<video|youtube\.com/embed`)

	page, err := storage.UpsertPage(ctx, pool, &storage.PageData{
		URL: p.URL, Title: title, MetaDescription: metaDesc,
		H1: h1, H2: h2, H3: h3,
		WordCount: wordCount, ImageCount: imgCount, VideoCount: vidCount,
	})
	if err != nil {
		log.Printf("extractor: upsert %s: %v", p.URL, err)
		return nil
	}

	kws, _ := storage.ListKeywords(ctx, pool, p.ProjectID, storage.ListOpts{Limit: 200})
	var keywordID string
	for _, k := range kws {
		if k.Keyword == p.Keyword {
			keywordID = k.ID
			break
		}
	}
	if keywordID != "" && page != nil {
		storage.InsertKeywordPage(ctx, pool, &storage.KeywordPage{
			KeywordID: keywordID, PageID: page.ID, Position: &p.Position,
		})
	}

	log.Printf("extractor: %s → %q words=%d h1=%d h2=%d", p.URL[:min(50, len(p.URL))], title, wordCount, len(h1), len(h2))

	// LLM topic extraction (best-effort, non-blocking via goroutine)
	go func() {
		tctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		h2text := strings.Join(h2, ". ")
		if len(h2text) > 500 {
			h2text = h2text[:500]
		}
		topics, err := llm.ExtractTopics(tctx, llmClient, title, strings.Join(h1, ". "), h2text, metaDesc)
		if err != nil {
			log.Printf("extractor: LLM topics %s: %v", p.URL[:min(40, len(p.URL))], err)
			return
		}
		log.Printf("extractor: topics for %s: %s", p.URL[:min(40, len(p.URL))], topics)
	}()

	val, _ := projectCount.LoadOrStore(p.ProjectID, new(atomic.Int32))
	val.(*atomic.Int32).Add(1)

	return nil
}

func download(rawURL string) (string, error) {
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

var (
	reTitle = regexp.MustCompile(`(?i)<title[^>]*>([^<]+)</title>`)
	reMeta  = regexp.MustCompile(`(?i)<meta[^>]+name=["']description["'][^>]+content=["']([^"']+)["']`)
	reMeta2 = regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']+)["'][^>]+name=["']description["']`)
)

func extractTitle(html string) string {
	m := reTitle.FindStringSubmatch(html)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractMeta(html string) string {
	m := reMeta.FindStringSubmatch(html)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	m = reMeta2.FindStringSubmatch(html)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func extractHeadings(html, tag string) []string {
	re := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[^>]*>([^<]*)</%s>`, tag, tag))
	var out []string
	for _, m := range re.FindAllStringSubmatch(html, 20) {
		if len(m) > 1 {
			if t := strings.TrimSpace(m[1]); t != "" {
				out = append(out, t)
			}
		}
	}
	return out
}

func countMatches(html, pattern string) int {
	return len(regexp.MustCompile(fmt.Sprintf(`(?i)%s`, pattern)).FindAllString(html, -1))
}

func stripTags(html string) string { return regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, " ") }

func countWords(text string) int { return len(strings.Fields(text)) }
