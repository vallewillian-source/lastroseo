# Storage Layer — AGENTS.md

> **Shared PostgreSQL schema, models, and CRUD for all services.**

## Role
Not a standalone service — a Go library imported by all services via `go.work` replace directive. Provides pgxpool connection, migrations, models, and query functions.

## Schema

| Table | Type | Purpose |
|-------|------|---------|
| `projects` | Standard | Client projects (name, business_desc) |
| `keywords` | Standard | All discovered keywords (seeds + autocomplete + serp_crawl) |
| `keyword_metrics` | **Hypertable** | Time-series: heat_score, serp_position |
| `clusters` | Standard | Named keyword clusters with intent |
| `serp_results` | **Hypertable** | Time-series: SERP snapshots (position, URL, title, snippet) |
| `page_data` | Standard | HTML extraction: titles, headings, word counts, topics |
| `keyword_pages` | Standard | Inverted index: keyword ↔ page |
| `jobs` | Standard | Job metadata (Asynq uses Redis, we mirror status in PG) |
| `content_gaps` | Standard | LLM-detected content gaps per keyword |

## Hypertables (TimescaleDB)
- `keyword_metrics`: partitioned by `timestamp`
- `serp_results`: partitioned by `crawled_at`

## Key models (Go structs)

```go
type Project struct {
    ID, Name, BusinessDesc string
    CreatedAt, UpdatedAt time.Time
}

type Keyword struct {
    ID, ProjectID, Keyword string
    IsSeed bool
    ClusterID, ClusterName *string
    Intent, Source *string
}

type SERPResult struct {
    ID, KeywordID string
    Position int
    URL, Title, Snippet string
    CrawledAt time.Time
}

type Cluster struct {
    ID, ProjectID, Name string
    Intent *string
    KeywordCount int
}

type ContentGap struct {
    ID, ProjectID, KeywordID, Keyword, Gaps string
}
```

## Key queries

| Function | File | Purpose |
|----------|------|---------|
| `NewPool` | `pool.go` | PG connection pool from env |
| `Migrate` | `migrate.go` | Run embedded SQL migrations |
| `ListKeywords` | `keywords.go` | Paginated keywords for a project |
| `UpdateKeywordCluster` | `keywords.go` | Set cluster_id + cluster_name + intent |
| `InsertSERPResult` | `metrics.go` | Single SERP row insert |
| `ListSERPResultsByProject` | `metrics.go` | JOIN serp_results + keywords |
| `GetClustersByProject` | `clusters.go` | All clusters for a project |
| `GetContentGaps` | `gaps.go` | All content gaps for a project |

## Migrations
- `001_init.sql` — full initial schema
- `002_gaps.sql` — content_gaps table

## Batch Writer
`SERPBatchWriter` in `batch.go` — buffers SERP results and flushes every 5s or 1000 records using COPY protocol.

## Dependencies
- `github.com/jackc/pgx/v5` — connection pool + COPY
- TimescaleDB extension (auto-loaded in migration)
