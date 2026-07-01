# Analytics & Clustering (analytics-svc)

Processes keyword data through 4 pipelines: semantic clustering (Union-Find + K-D Tree), intent classification, heat score tracking, content gap detection. Uses local Ollama LLMs with DeepSeek API fallback.

## Quick start

```bash
go build -o analytics-svc .
OLLAMA_URL=http://localhost:11434 REDIS_ADDR=localhost:6379 ./analytics-svc
```

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_URL` | `http://localhost:11434` | Ollama host |
| `DEEPSEEK_API_KEY` | — | Optional fallback |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `POSTGRES_*` | — | PostgreSQL connection |
