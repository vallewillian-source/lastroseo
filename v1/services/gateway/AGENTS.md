# Gateway Service — AGENTS.md

> **REST API entrypoint. Single producer of Asynq jobs.**

## Role
Receives all external requests (from frontend or CLI), validates input, enqueues Asynq jobs, and serves read-only queries from PostgreSQL.

## Tech stack
- **Chi router** — middleware: Logger, Recoverer, Timeout (30s)
- **Asynq** — Redis task queue, producer role
- **pgxpool** — PostgreSQL connection pool
- **go-redis** — Redis client for health checks

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/health` | — | Liveness (always 200) |
| GET | `/readiness` | — | Startup probe (PG + Redis pings) |
| GET | `/services` | — | Services health (PG, Redis, SearXNG) |
| POST | `/api/v1/projects` | — | Create project, enqueue KEYWORD_RESEARCH, return 202 |
| GET | `/api/v1/projects` | — | List all projects |
| GET | `/api/v1/projects/{id}` | — | Project detail |
| GET | `/api/v1/projects/{id}/keywords` | — | Keywords with cluster + intent |
| GET | `/api/v1/projects/{id}/serp` | — | SERP results grouped by keyword |
| GET | `/api/v1/projects/{id}/clusters` | — | Clusters with keyword counts |
| GET | `/api/v1/projects/{id}/gaps` | — | Content gaps from LLM |
| GET | `/api/v1/jobs/{id}` | — | Job status |

All `/api/v1/` routes also registered under `/api/` prefix for nginx proxy support.

## Job orchestration

| Job Type | Queue | Target | Trigger |
|----------|-------|--------|---------|
| `KEYWORD_RESEARCH` | default | google-svc | Project creation |
| `SERP_CRAWL` | default | crawler-svc | google-svc |
| `CONTENT_EXTRACTION` | low | extractor-svc | crawler-svc (per URL) |
| `ANALYTICS` | default | analytics-svc | crawler-svc (per project) |

## Flow
```
POST /api/v1/projects
  → validate JSON
  → DB: insert project + seed keywords
  → Asynq: enqueue KEYWORD_RESEARCH {project_id, name, business_desc, seed_keywords}
  → DB: create job record (status: PENDING)
  → 202 Accepted {project_id, job_id, asynq_id}
```

## Configuration
Environment variables:
- `POSTGRES_HOST`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`
- `REDIS_ADDR` (default: `redis:6379`)
- `SEARXNG_URL` (default: `http://searxng:8080`)
- Server listens on `:8080` (mapped to `:8085` externally)

## Dependencies
- **Upstream**: PostgreSQL, Redis, SearXNG (health check only)
- **Downstream**: google-svc, crawler-svc, extractor-svc, analytics-svc (via Asynq)
