# SERP Crawler (crawler-svc)

Scrapes Google SERPs via SearXNG. For each keyword, collects top 10 organic results and stores historical snapshots in TimescaleDB.

## Quick start

```bash
go build -o crawler-svc .
SEARXNG_URL=http://localhost:8080 REDIS_ADDR=localhost:6379 ./crawler-svc
```

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `SEARXNG_URL` | `http://localhost:8080` | SearXNG instance |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `POSTGRES_*` | — | PostgreSQL connection |
