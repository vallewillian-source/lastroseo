# Storage Layer

PostgreSQL 16 + TimescaleDB schema and batch-write utilities for LastroSEO V1.

## Schema
- **7 standard tables**: projects, seed_keywords, keywords, clusters, page_data, keyword_pages, jobs
- **2 hypertables**: keyword_metrics, serp_results
- **Full-text indexes**: GIN on keywords (Portuguese tsvector)
- **Inverted index indexes**: B-tree on keyword_id + page_id

## Migrations
Applied in order from `migrations/` directory. Initial schema: `migrations/001_init.sql`.

## Batch Writer
Buffers records, flushes every 10k records or 5 seconds via PostgreSQL COPY protocol.
