# Gateway Service (gateway-svc)

REST API entrypoint for LastroSEO V1. Accepts project submissions, enqueues background jobs via Asynq (Redis), and serves keyword research results.

## Quick start

```bash
# Build
cd services/gateway
go build -o gateway-svc .

# Run (requires PG + Redis)
POSTGRES_HOST=localhost REDIS_ADDR=localhost:6379 ./gateway-svc

# Health check
curl http://localhost:8080/health
```

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness probe |
| `GET` | `/readiness` | Startup probe |
| `POST` | `/api/v1/projects` | Create project |
| `GET` | `/api/v1/projects/{id}` | Project detail |
| `GET` | `/api/v1/projects/{id}/keywords` | Keyword list |
| `GET` | `/api/v1/jobs/{id}` | Job status |

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_USER` | `lastroseo` | Database user |
| `POSTGRES_PASSWORD` | `secret` | Database password |
| `POSTGRES_DB` | `lastroseo` | Database name |
| `REDIS_ADDR` | `localhost:6379` | Redis address |

## Docker

```bash
docker build -t lastroseo-gateway .
docker run -p 8081:8080 --env-file ../../.env lastroseo-gateway
```
