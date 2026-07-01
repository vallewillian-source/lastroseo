# LastroSEO V1 — AGENTS.md

> **Implemented architecture. Reflects what is running, not plans.**

## Desktop environment

| Item | Detail |
|------|--------|
| **Host** | Ubuntu desktop `192.168.0.23` |
| **SSH** | `ssh willianvalle@192.168.0.23` (key-based, no password) |
| **Project path** | `/home/willianvalle/lastroseo/v1` |
| **Local workspace** | `/Users/vallewillian/workspace/lastro-seo/v1` (Mac) |
| **Ollama** | Runs on **host** (not container), model `gemma4:e4b` |
| **Ollama URL from containers** | `http://172.17.0.1:11434` (Docker host gateway) |
| **Docker** | Engine + Compose v2 (`docker compose`, not `docker-compose`) |
| **DB credentials** | `lastroseo` / `secret` / `lastroseo` (dev only) |

### Ollama setup (desktop)

```bash
# Install (once)
curl -fsSL https://ollama.com/install.sh | sh
ollama pull gemma4:e4b

# Verify
ollama list                          # should list gemma4:e4b
curl http://localhost:11434/api/tags # API check

# Make accessible from Docker containers:
sudo systemctl edit ollama
# Add: Environment="OLLAMA_HOST=0.0.0.0"
sudo systemctl restart ollama
```

### Network map

```
Desktop 192.168.0.23
├── Ollama (host)     :11434
└── Docker
    ├── postgres      :5433 → :5432
    ├── redis         :6380 → :6379
    ├── searxng       :8083 → :8080
    ├── gateway-svc   :8085 → :8080
    ├── frontend      :3011 → :80   (nginx → gateway:8080)
    ├── google-svc    (Asynq worker, no port)
    ├── crawler-svc   (Asynq worker)
    ├── extractor-svc (Asynq worker + Ollama)
    ├── analytics-svc (Asynq worker + Ollama)
    └── storage-svc   :8082 → :8080
```

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
| `gateway-svc` | REST API + Competitor Inspect (Gemma4) | `services/gateway/AGENTS.md` |
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
10. Frontend: Inspect Keywords → Gateway fetches competitor HTML → Gemma4 extracts keywords
```

## Ports

| Service | External | Internal |
|---------|----------|----------|
| Gateway | 8085 | 8080 |
| Frontend | 3011 | 80 |
| PostgreSQL | 5433 | 5432 |
| Redis | 6380 | 6379 |
| SearXNG | 8083 | 8080 |
| Ollama (host) | — | 11434 |

## Quick start

```bash
# ---- Desktop (192.168.0.23) ----

# 1. Garantir Ollama rodando no host
ollama list
sudo systemctl status ollama

# 2. Infra
cd /home/willianvalle/lastroseo/v1
cp .env.example .env   # só precisa se não existir
docker compose up -d    # postgres, redis, searxng

# 3. Apps (aguardar infra healthy)
docker compose --profile app build --network host
docker compose --profile app up -d

# 4. Verificar
curl http://localhost:8085/health
curl http://localhost:3011        # frontend

# ---- Local (Mac) ----
# Acessar via browser: http://192.168.0.23:3011
# SSH para debug: ssh willianvalle@192.168.0.23
```

## Dev rules

1. Never commit `.env`
2. Each service has own `go.mod` → linked via `go.work`
3. Replace directives: `storage` and `pkg/*` are local
4. All services use `FROM scratch` Docker images (~8 MB)
5. Build with `--network host` to bypass Docker DNS issues
6. DNS: services use `8.8.8.8, 1.1.1.1` (no systemd-resolved)
7. Ollama host: `host.docker.internal:11434` → Docker `extra_hosts` needed on Linux

## Deploy & Debug

### Rebuild de um serviço específico

```bash
ssh willianvalle@192.168.0.23
cd /home/willianvalle/lastroseo/v1

# Gateway (exemplo)
docker compose --profile app build --network host gateway-svc
docker compose --profile app up -d gateway-svc

# Frontend
cd frontend && npm run build       # localmente ou no desktop
docker compose --profile app build --network host frontend
docker compose --profile app up -d frontend
```

### Logs

```bash
docker compose logs -f --tail=50 gateway-svc
docker compose logs -f --tail=50 extractor-svc
docker compose logs -f --tail=50 analytics-svc
```

### Checar saúde

```bash
curl http://192.168.0.23:8085/health       # gateway
curl http://192.168.0.23:8085/services     # PG + Redis + SearXNG
curl http://192.168.0.23:8083/health       # SearXNG
curl http://192.168.0.23:11434/api/tags    # Ollama + modelos
docker compose ps                           # containers
```

### Problemas comuns

| Sintoma | Causa provável | Solução |
|---------|---------------|---------|
| `gemma: connect failed` | Ollama não aceita conexões externas | `OLLAMA_HOST=0.0.0.0` no systemd |
| `gemma: status 404 model not found` | Modelo não baixado | `ollama pull gemma4:e4b` |
| `SERP: 0 resultados` | SearXNG bloqueado/offline | `curl http://192.168.0.23:8083/health` |
| `Asynq: NOBLIST` | Redis com chaves pendentes | `docker compose restart redis` |
| `fetch failed: context deadline exceeded` | Site lento/bloqueado | Aumentar timeout em `competitors.go` |
| Frontend branco | JS quebrou no build | `npm run build` novamente, limpar cache |
| `dial tcp: lookup` nos containers | DNS do Docker quebrou | Containers usam `8.8.8.8, 1.1.1.1` |

### Notas sobre builds

- **Sempre use `--network host`** no build: o Docker na versão usada tem problemas de DNS em bridge network.
- Imagens Go usam `FROM scratch` (~8 MB), então precisam de `CGO_ENABLED=0`.
- O `go.work` na raiz de `v1/` linka todos os módulos locais (`storage`, `pkg/*`, serviços).
- Ao alterar `storage/` ou `pkg/*`, rebuild de **todos** os serviços que os importam.
- Frontend: build local (`npm run build`) gera `dist/`, o Dockerfile só copia esse diretório.
- `.env` **nunca** deve ser commitado. O `.env.example` é o template.
