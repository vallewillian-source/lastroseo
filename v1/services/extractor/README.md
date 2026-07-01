# Content Extractor (extractor-svc)

Downloads and parses HTML from URLs found in SERPs. Extracts metadata (title, meta, headings, word count) and computes TF-IDF keywords. Builds a double inverted index (URL ↔ Keywords).

## Quick start

```bash
go build -o extractor-svc .
OLLAMA_URL=http://localhost:11434 REDIS_ADDR=localhost:6379 ./extractor-svc
```

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_URL` | `http://localhost:11434` | Ollama for content gap detection |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `POSTGRES_*` | — | PostgreSQL connection |
