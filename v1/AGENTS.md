# LastroSEO V1 — AGENTS.md

> **Implemented architecture. Reflects what is running, not plans.**

## Architecture

```
                    ┌──────────────┐
 User / Frontend ─→ │  Gateway     │ Chi REST :8085 (internal :8080)
                    │  (producer)  │ Asynq enqueue
                    └──────┬───────┘
                           │ Redis / Asynq
           ┌───────────────┼───────────────┬──────────────┐
    ┌──────▼──────┐ ┌─────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
    │ Google      │ │ Crawler    │ │ Extractor   │ │ Analytics  │
    │ Fetcher     │ │ SearXNG    │ │ regex + LLM │ │ Jaccard +  │
    │ Autocomplete│ │            │ │ topics      │ │ Gemma      │
    └──────┬──────┘ └─────┬──────┘ └──────┬──────┘ └─────┬──────┘
           │              │               │               │
           └──────────────┴───────────────┴───────────────┘
                           │
                    ┌──────▼──────┐
                    │  Storage    │
                    │  PG 16 +    │
                    │  TimescaleDB│
                    └─────────────┘
```

Six Go microservices + React frontend, Docker Compose on Ubuntu desktop.

## Core technology decisions

| Choice | Why |
|--------|-----|
| **Go 1.23–1.25** | Static scratch binaries (~8 MB), goroutines |
| **Chi router** | Lightweight HTTP, no framework lock-in |
| **Asynq (Redis)** | Task queue: retry, dead-letter, priorities |
| **PostgreSQL 16 + TimescaleDB** | Relational + time-series in one instance |
| **Redis 7** | Asynq broker + cache |
| **SearXNG** | Meta search, self-hosted, no API key |
| **Gemma 4B (Ollama)** | Local LLM for intent, cluster naming, gaps, topics |
| **React + React Spectrum** | Frontend with Adobe design system |
| **Nginx** | SPA proxy + API gateway reverse proxy |

## What is NOT implemented (by design)

- **gRPC** — Asynq + direct PG queries are simpler for this scale
- **OpenTelemetry/Prometheus** — stdout logging is sufficient
- **K-D Tree / embeddings** — Jaccard clustering works for < 500 keywords
- **Roaring Bitmap / Trie / Bloom Filter** — sync.Map + PG DISTINCT suffice
- **Google Ads API / Trends / Reddit** — only free Autocomplete scraping
- **DeepSeek API** — Ollama Gemma 4B handles everything locally

## Services

| Service | Type | Doc |
|---------|------|-----|
| `gateway-svc` | REST API | `services/gateway/AGENTS.md` |
| `google-svc` | Asynq worker | `services/google-fetcher/AGENTS.md` |
| `crawler-svc` | Asynq worker | `services/crawler/AGENTS.md` |
| `extractor-svc` | Asynq worker | `services/extractor/AGENTS.md` |
| `analytics-svc` | Asynq worker | `services/analytics/AGENTS.md` |
| `storage` | Shared library | `services/storage/AGENTS.md` |

## Pipeline (Phase 1 — complete)

```
1. POST /api/v1/projects → Gateway → Asynq KEYWORD_RESEARCH
2. google-svc → Autocomplete scraping (~8 suggestions/seed)
3. google-svc → Asynq SERP_CRAWL (all keywords)
4. crawler-svc → SearXNG (top 10/kw, 2s rate limit, 10 workers)
5. crawler-svc → Asynq CONTENT_EXTRACTION (per unique URL)
6. crawler-svc → Asynq ANALYTICS (1 per project)
7. extractor-svc → regex HTML parse + Gemma topic extraction
8. analytics-svc → Jaccard clustering + regex/LLM intent + Gemma naming + gap detection
9. Frontend polls: keywords, SERP, clusters, gaps
```

## Ports

| Service | External | Internal |
|---------|----------|----------|
| Gateway | 8085 | 8080 |
| Frontend | 3011 | 80 |
| PostgreSQL | 5433 | 5432 |
| Redis | 6380 | 6379 |
| SearXNG | 8083 | 8080 |

## Quick start

```bash
cp .env.example .env
docker-compose up -d                         # infra (postgres, redis, searxng)
docker-compose --profile app up -d           # all app services + frontend
open http://192.168.0.23:3011                # frontend
curl http://192.168.0.23:8085/health         # API
```

## Dev rules

1. Never commit `.env`
2. Each service has own `go.mod` → linked via `go.work`
3. Replace directives: `storage` and `pkg/*` are local
4. All services use `FROM scratch` Docker images (~8 MB)
5. Build with `--network host` to bypass Docker DNS issues
6. DNS: services use `8.8.8.8, 1.1.1.1` (no systemd-resolved)
7. Ollama host: `host.docker.internal:11434` → Docker `extra_hosts` needed on Linux
