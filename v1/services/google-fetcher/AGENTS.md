# Google Data Fetcher — AGENTS.md

> **Asynq worker. Keyword expansion via Google Autocomplete scraping.**

## Role
Consumes `KEYWORD_RESEARCH` jobs. Expands seed keywords via Google Autocomplete scraping (free, no API key). Stores discovered keywords in PostgreSQL, then enqueues `SERP_CRAWL` for all keywords.

## Pipeline (per KEYWORD_RESEARCH job)

```
1. Load payload: {project_id, name, business_desc, seed_keywords}
2. Insert seeds as is_seed=true, source="seed"
3. For each seed keyword: call Google Autocomplete
   GET suggestqueries.google.com/complete/search?client=firefox&q={keyword}
   Parse JSON → extract ~8 suggestions per seed
4. Dedup: map[string]bool merges seeds + suggestions
5. Insert new keywords as source="autocomplete"
6. Enqueue SERP_CRAWL with ALL keywords (seeds + expanded)
```

## Autocomplete scraping
```
URL: https://suggestqueries.google.com/complete/search?client=firefox&q={keyword}
Response: ["keyword", ["suggestion1", "suggestion2", ...]]
```
No API key required. Rate limit handled by HTTP client timeout (10s).

## Worker config
```go
asynq.Config{Concurrency: 5, Queues: map[string]int{"default": 3, "low": 1}}
```

## Output
- `keywords` table — seeds + autocomplete suggestions
- Enqueues: `SERP_CRAWL` with full keyword list

## What is NOT implemented
- Google Ads Keyword Planner (requires API key)
- Google Trends scraping (not needed for MVP)
- Reddit/Quora scraping (future)
- Trie caching (PG DISTINCT is sufficient for current scale)

## Dependencies
- `github.com/hibiken/asynq`
- `github.com/jackc/pgx/v5`
- `github.com/lastroseo/services/storage`
