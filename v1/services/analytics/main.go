// Analytics & Clustering (analytics-svc)
//
// Asynq worker that consumes ANALYTICS jobs.
// Pipelines: Jaccard clustering (Union-Find), regex + Gemma LLM intent classification,
// semantic cluster naming (Gemma), heat score from SERP rankings, content gap detection.
package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	llm "github.com/lastroseo/pkg/llm"
	storage "github.com/lastroseo/services/storage"
)

var (
	pool      *pgxpool.Pool
	redisAddr = getEnv("REDIS_ADDR", "localhost:6379")
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
		log.Fatalf("analytics: pg: %v", err)
	}
	defer pool.Close()
	log.Println("   PostgreSQL connected")

	llmClient = llm.NewClient(os.Getenv("OLLAMA_URL"), "gemma4:e4b")
	log.Printf("   LLM: gemma4:e4b @ %s", os.Getenv("OLLAMA_URL"))

	redisCli := goredis.NewClient(&goredis.Options{Addr: redisAddr})
	defer redisCli.Close()

	asynqSrv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 3, Queues: map[string]int{"default": 2, "low": 1}},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc("ANALYTICS", handleAnalytics)

	go func() {
		log.Println("🧠 Analytics — clustering, intent, heat")
		if err := asynqSrv.Run(mux); err != nil {
			log.Fatalf("analytics: asynq: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("analytics: shutting down")
	asynqSrv.Shutdown()
}

type analyticsPayload struct {
	ProjectID string `json:"project_id"`
}

func handleAnalytics(ctx context.Context, t *asynq.Task) error {
	var p analyticsPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	log.Printf("analytics: ANALYTICS project=%s", p.ProjectID)

	// 1. Load all keywords for this project
	kws, err := storage.ListKeywords(ctx, pool, p.ProjectID, storage.ListOpts{Limit: 500})
	if err != nil {
		return err
	}
	if len(kws) == 0 {
		log.Printf("analytics: project=%s no keywords", p.ProjectID)
		return nil
	}
	log.Printf("analytics: loaded %d keywords", len(kws))

	// 2. Cluster keywords by Jaccard similarity
	clusters := clusterKeywords(kws)
	log.Printf("analytics: found %d clusters", len(clusters))

	// 3. Classify intent per keyword (regex first, LLM fallback for unknowns)
	for i := range kws {
		intent := classifyIntent(kws[i].Keyword)
		if intent == "" {
			// Regex didn't match — try LLM
			ictx, cancel := context.WithTimeout(ctx, 30*time.Second)
			if llmIntent, err := llm.ClassifyIntent(ictx, llmClient, kws[i].Keyword, ""); err == nil {
				intent = llmIntent
				log.Printf("analytics: LLM intent %q → %s", kws[i].Keyword, intent)
			} else {
				intent = "informacional"
				log.Printf("analytics: LLM intent failed for %q: %v", kws[i].Keyword, err)
			}
			cancel()
		}
		kws[i].Intent = strPtr(intent)
	}

	// 4. Compute heat scores from SERP data
	computeHeatScores(ctx, pool, p.ProjectID, kws)

	// 5. Persist clusters + update keywords (LLM naming)
	for _, c := range clusters {
		// Try LLM cluster naming, fallback to keyword-based
		cctx, ccancel := context.WithTimeout(ctx, 15*time.Second)
		clusterName := c.Name
		if llmName, err := llm.NameCluster(cctx, llmClient, clusterKeywordsText(c.Keywords, kws)); err == nil {
			clusterName = llmName
			log.Printf("analytics: LLM cluster name: %q", clusterName)
		} else {
			log.Printf("analytics: LLM cluster name failed: %v", err)
		}
		ccancel()

		cluster, err := storage.CreateCluster(ctx, pool, &storage.Cluster{
			ProjectID:    p.ProjectID,
			Name:         clusterName,
			Intent:       dominantIntent(c.Keywords, kws),
			KeywordCount: len(c.Keywords),
		})
		if err != nil {
			log.Printf("analytics: create cluster: %v", err)
			continue
		}
		for _, kw := range c.Keywords {
			intent := getIntent(kw, kws)
			if err := storage.UpdateKeywordCluster(ctx, pool, kw, cluster.ID, cluster.Name, intent); err != nil {
				log.Printf("analytics: update keyword %q: %v", kw, err)
			}
		}
	}

	// 6. Content gap detection (per keyword with SERP data)
	detectGaps(ctx, p.ProjectID, kws)

	log.Printf("analytics: done project=%s clusters=%d keywords=%d", p.ProjectID, len(clusters), len(kws))
	return nil
}

// ── Clustering ─────────────────────────────────────────────────

type jaccardCluster struct {
	Name     string
	Keywords []string // keyword IDs
}

func clusterKeywords(kws []storage.Keyword) []jaccardCluster {
	type idxPair struct{ i, j int }
	type pairScore struct {
		pair  idxPair
		score float64
	}

	n := len(kws)
	tokenSets := make([]map[string]struct{}, n)
	for i, kw := range kws {
		tokenSets[i] = tokenize(kw.Keyword)
	}

	// Compute all pairwise Jaccard similarities above threshold
	var edges []pairScore
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			sim := jaccard(tokenSets[i], tokenSets[j])
			if sim > 0.35 {
				edges = append(edges, pairScore{pair: idxPair{i, j}, score: sim})
			}
		}
	}

	// Sort by similarity descending
	sort.Slice(edges, func(a, b int) bool { return edges[a].score > edges[b].score })

	// Union-Find
	parent := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	var find func(x int) int
	find = func(x int) int {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}

	for _, e := range edges {
		union(e.pair.i, e.pair.j)
	}

	// Group by root
	groups := map[int][]int{}
	for i := 0; i < n; i++ {
		r := find(i)
		groups[r] = append(groups[r], i)
	}

	// Build result
	var result []jaccardCluster
	for _, indices := range groups {
		c := jaccardCluster{}
		for _, idx := range indices {
			c.Keywords = append(c.Keywords, kws[idx].ID)
		}
		c.Name = clusterNameFromIndices(indices, kws)
		result = append(result, c)
	}

	// Any keyword not clustered goes solo
	clustered := map[string]bool{}
	for _, c := range result {
		for _, id := range c.Keywords {
			clustered[id] = true
		}
	}
	for _, kw := range kws {
		if !clustered[kw.ID] {
			result = append(result, jaccardCluster{
				Name:     longestWord(kw.Keyword),
				Keywords: []string{kw.ID},
			})
		}
	}

	return result
}

func tokenize(s string) map[string]struct{} {
	s = strings.ToLower(s)
	// Split on non-letters
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		if len(w) > 2 { // ignore very short words
			set[w] = struct{}{}
		}
	}
	return set
}

func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	intersection := 0
	for w := range a {
		if _, ok := b[w]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func clusterNameFromIndices(indices []int, kws []storage.Keyword) string {
	wordFreq := map[string]int{}
	for _, idx := range indices {
		for w := range tokenize(kws[idx].Keyword) {
			wordFreq[w]++
		}
	}
	best := ""
	bestN := 0
	for w, n := range wordFreq {
		if n > bestN || (n == bestN && len(w) > len(best)) {
			best = w
			bestN = n
		}
	}
	return best
}

func longestWord(kw string) string {
	words := strings.FieldsFunc(kw, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	best := ""
	for _, w := range words {
		if len(w) > len(best) {
			best = w
		}
	}
	return best
}

// ── Intent Classification ─────────────────────────────────────

var intentRules = []struct {
	pattern *regexp.Regexp
	intent  string
}{
	{regexp.MustCompile(`\b(como|o que .|quem|onde|por que|para que|qual|quais|guia|tutorial|aprender|dicas|passo a passo|melhor forma|melhor maneira|exemplos|benefícios|vantagens|desvantagens|funciona|significado|definição|conceito)\b`), "informacional"},
	{regexp.MustCompile(`\b(comprar|preço|preços|barato|orçamento|orçar|contratar|plano|planos|assinatura|mensalidade|custo|valor|quanto custa|preço de|tabela de preço|promoção|desconto|frete grátis|entrega)\b`), "transacional"},
	{regexp.MustCompile(`\b(vs|versus|comparação|comparativo|alternativa|concorrente|melhor |top |ranking|review|avaliação|análise|opinião|vale a pena|prós e contras|diferencial|diferenciais)\b`), "comercial"},
	{regexp.MustCompile(`\b(site|login|entrar|acessar|baixar|download|app|aplicativo|telefone|contato|endereço|horário|atendimento|suporte|portal|área do cliente|consulta|2 via|segunda via)\b`), "navegacional"},
}

func classifyIntent(kw string) string {
	kw = strings.ToLower(kw)
	for _, rule := range intentRules {
		if rule.pattern.MatchString(kw) {
			return rule.intent
		}
	}
	return "" // trigger LLM fallback
}

func getIntent(kwID string, kws []storage.Keyword) string {
	for _, k := range kws {
		if k.ID == kwID && k.Intent != nil {
			return *k.Intent
		}
	}
	return "informacional"
}

func dominantIntent(kwIDs []string, kws []storage.Keyword) *string {
	counts := map[string]int{}
	for _, id := range kwIDs {
		for _, k := range kws {
			if k.ID == id && k.Intent != nil {
				counts[*k.Intent]++
				break
			}
		}
	}
	best := ""
	bestN := 0
	for intent, n := range counts {
		if n > bestN {
			best = intent
			bestN = n
		}
	}
	if best != "" {
		return &best
	}
	return nil
}

// ── Heat Score ─────────────────────────────────────────────────

func computeHeatScores(ctx context.Context, pool *pgxpool.Pool, projectID string, kws []storage.Keyword) {
	// For each keyword, compute position-based heat: 1st = 100, 10th = 10
	for _, kw := range kws {
		results, err := storage.GetSERPHistory(ctx, pool, kw.ID, time.Now().Add(-24*time.Hour), 10)
		if err != nil || len(results) == 0 {
			continue
		}
		// Average of inverse position scores
		var sum float64
		for _, r := range results {
			sum += math.Max(0, 100-float64((r.Position-1)*10))
		}
		heat := sum / float64(len(results))

		// Store as keyword metric
		_, _ = pool.Exec(ctx,
			`INSERT INTO keyword_metrics (keyword_id, timestamp, heat_score, serp_position)
			 VALUES ($1, NOW(), $2, $3)`,
			kw.ID, heat, results[0].Position,
		)
	}
}

func strPtr(s string) *string { return &s }

// clusterKeywordsText returns the keyword strings for a set of keyword IDs.
func clusterKeywordsText(ids []string, kws []storage.Keyword) []string {
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	var texts []string
	for _, kw := range kws {
		if idSet[kw.ID] {
			texts = append(texts, kw.Keyword)
		}
	}
	return texts
}

// detectGaps runs content gap detection for up to 3 top keywords.
func detectGaps(ctx context.Context, projectID string, kws []storage.Keyword) {
	if len(kws) == 0 {
		return
	}
	// Process top 3 keywords (by SERP data volume)
	count := 0
	for _, kw := range kws {
		if count >= 3 {
			break
		}
		results, err := storage.GetSERPHistory(ctx, pool, kw.ID, time.Now().Add(-7*24*time.Hour), 5)
		if err != nil || len(results) == 0 {
			continue
		}

		// Collect H2s from competitor pages (approximation: use snippets as topics)
		var topics []string
		for _, r := range results {
			if r.Snippet != "" {
				topics = append(topics, r.Title)
			}
		}
		if len(topics) < 3 {
			continue
		}

		gctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		gapText, err := llm.DetectContentGaps(gctx, llmClient, kw.Keyword, topics)
		cancel()
		if err != nil {
			log.Printf("analytics: gap detection %q: %v", kw.Keyword, err)
			continue
		}

		// Store gap
		_, err = pool.Exec(ctx,
			`INSERT INTO content_gaps (project_id, keyword_id, keyword, gaps)
			 VALUES ($1, $2, $3, $4)`,
			projectID, kw.ID, kw.Keyword, gapText,
		)
		if err != nil {
			log.Printf("analytics: insert gap %q: %v", kw.Keyword, err)
		} else {
			log.Printf("analytics: gap detected for %q: %s", kw.Keyword, gapText[:min(60, len(gapText))])
		}
		count++
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
