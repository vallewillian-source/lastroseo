# Google Data Fetcher (google-svc)

Asynq worker that collects keyword data from Google sources and enriches the keyword database.

## Sources
- **Google Ads Keyword Planner** — volume, CPC, competition (optional, needs API key)
- **Google Autocomplete** — 26× suggestions per seed keyword
- **Google Trends** — interest-over-time + rising queries
- **Reddit/Quora** — discussion scraping (future)

## Quick start

```bash
go build -o google-svc .
./google-svc  # connects to PG + Redis, waits for Asynq jobs
```

## Environment

| Variable | Required | Description |
|----------|----------|-------------|
| `POSTGRES_*` | Yes | PostgreSQL connection |
| `REDIS_ADDR` | Yes | Redis address |
| `GOOGLE_ADS_*` | No | Ads API credentials |

## Cache layers
- L1: In-memory (1h TTL)
- L2: Redis (24h TTL)
- L3: PostgreSQL (persistent)
