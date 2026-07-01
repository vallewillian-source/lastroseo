# SERP Crawler — AGENTS.md

> **Asynq worker. SearXNG client. 10 concurrent workers, 2s rate limit.**

## Role
Consumes `SERP_CRAWL` jobs. Queries SearXNG for each keyword (top 10 results), stores in `serp_results` hypertable, deduplicates URLs via FNV-1a hash in sync.Map, enqueues `CONTENT_EXTRACTION` per unique URL, and enqueues one `ANALYTICS` job per project when done.

## Pipeline (per SERP_CRAWL job)

```
1. Load payload: {project_id, keywords[]}
2. Upsert keywords (source="serp_crawl")
3. For each keyword (goroutine, max 5 concurrent):
   a. Wait for rate-limit ticker (2s)
   b. GET SearXNG /search?q={keyword}&format=json&language=pt-BR
   c. Store top 10 results in serp_results
   d. For each URL: check FNV-1a hash in sync.Map
      → new: enqueue CONTENT_EXTRACTION with {project_id, url, keyword, position}
      → seen: skip
4. After all keywords: enqueue ANALYTICS {project_id}
```

## Key implementation details

```go
// Rate limit
ticker = time.NewTicker(2 * time.Second)  // ~30 req/min

// URL dedup (in-memory, per process lifetime)
visited sync.Map  // uint64(FNV-1a) → struct{}

// Worker pool
asynq.Config{Concurrency: 10, Queues: map[string]int{"default": 3, "low": 1}}
// Max 5 concurrent goroutines via semaphore channel
```

## SearXNG integration
```
GET http://searxng:8080/search?q={keyword}&format=json&categories=general&language=pt-BR
Response: {results: [{url, title, content}]}
Timeout: 15s per request
```

## Output tables
- `serp_results` — one row per result (keyword_id, position, url, title, snippet, crawled_at)
- `keywords` — upsert with source="serp_crawl"
- Enqueues: CONTENT_EXTRACTION (per new URL), ANALYTICS (per project)

## Dependencies
- `github.com/hibiken/asynq`
- `github.com/jackc/pgx/v5`
- `github.com/lastroseo/services/storage`
- SearXNG (HTTP, no Go library needed)
