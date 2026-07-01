# Analytics & Clustering — AGENTS.md

> **Asynq worker. Jaccard clustering + regex/LLM intent + Gemma cluster naming + content gaps.**

## Role
Consumes `ANALYTICS` jobs. Runs 4 pipelines on keyword data: clustering, intent classification, heat scoring, content gap detection. Uses Gemma 4B via Ollama for semantic tasks (intent fallback, cluster naming, gap detection).

## Pipeline (per ANALYTICS job)

### 1. Keyword Clustering (Jaccard + Union-Find)
```
Input: all project keywords
  1. Tokenize each keyword (split on non-letters, min 3 chars)
  2. Compute pairwise Jaccard similarity (|A ∩ B| / |A ∪ B|)
  3. If Jaccard > 0.35 → Union(pair)
  4. After all pairs: each Union-Find root = one cluster
  5. Name clusters via Gemma LLM:
     Prompt: "Give a short name for this keyword group: [keywords]"
     Fallback: most frequent word (if LLM unavailable)
Output: clusters table + keyword.cluster_id updates
```

### 2. Intent Classification (Regex → LLM fallback)
```go
// Tier 1: Regex rules (Portuguese)
patterns := {
  "como|guia|tutorial|o que é|dicas|aprender" → informacional
  "preço|comprar|orçamento|plano|contratar"  → transacional
  "melhor|vs|review|alternativa|vale a pena"  → comercial
  "login|site|app|download|telefone|suporte"   → navegacional
}

// Tier 2: Gemma 4B (when no regex match)
Prompt: "Classify this keyword intent: '{keyword}'. Business: {desc}.
         Answer: informacional|transacional|comercial|navegacional"
Timeout: 30s. Default: "informacional" on failure.
```

### 3. Heat Score (SERP position proxy)
```
For each keyword:
  Load last 10 SERP results from 24h window
  Score = Σ max(0, 100 - (position-1)*10) / count
  Store in keyword_metrics (hypertable)
```

### 4. Content Gap Detection (Gemma 4B)
```
For top 3 keywords (by SERP data volume):
  1. Collect competitor page titles from top 5 SERP results
  2. Prompt Gemma:
     "Analyze competitor topics for '{keyword}'. List 3-5 content gaps.
      Competitor topics: [titles]"
  3. Store gaps in content_gaps table
Timeout: 30s.
```

## Key data structures

```go
// Union-Find with path compression
type uf struct {
    parent []int
}

// Jaccard edge for clustering
type pairScore struct {
    i, j int
    score float64
}

// Intent classification
type intentRule struct {
    pattern *regexp.Regexp
    intent  string
}
```

## LLM strategy

| Task | Model | Timeout | Fallback |
|------|-------|---------|----------|
| Intent (fallback) | Gemma 4B | 30s | "informacional" |
| Cluster naming | Gemma 4B | 15s | Most frequent word |
| Content gaps | Gemma 4B | 30s | Skip keyword |
| Page topics | Gemma 4B | 30s | Skip page |

All LLM calls go through `pkg/llm` → Ollama HTTP API at `OLLAMA_URL`.

## Output tables updated
- `clusters` — new rows with semantic names
- `keywords` — cluster_id, cluster_name, intent
- `keyword_metrics` — heat_score, serp_position
- `content_gaps` — LLM-detected gaps per keyword

## What is NOT implemented
- K-D Tree / embeddings → Jaccard is sufficient for < 500 keywords
- DeepSeek API fallback → Gemma handles everything locally
- Sliding window trend analysis → simple position-based heat score

## Dependencies
- `github.com/hibiken/asynq`
- `github.com/jackc/pgx/v5`
- `github.com/redis/go-redis/v9`
- `github.com/lastroseo/pkg/llm` — Gemma 4B client
- `github.com/lastroseo/services/storage`
- Ollama (host desktop, model: gemma4:e4b)
