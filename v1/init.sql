-- LastroSEO V1 — Database Schema
-- PostgreSQL 16 + TimescaleDB

-- ── Projects ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS projects (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    business_desc   TEXT,
    target_audience TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ── Seed Keywords ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS seed_keywords (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    keyword     TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, keyword)
);

-- ── Keywords (all discovered) ────────────────────────────────
CREATE TABLE IF NOT EXISTS keywords (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    keyword         TEXT NOT NULL,
    is_seed         BOOLEAN DEFAULT FALSE,
    cluster_id      UUID,
    cluster_name    TEXT,
    intent          VARCHAR(20),
    source          VARCHAR(50),
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, keyword)
);

CREATE INDEX IF NOT EXISTS idx_keywords_search ON keywords
    USING gin(to_tsvector('portuguese', keyword));

-- ── Keyword Metrics (time series) ────────────────────────────
CREATE TABLE IF NOT EXISTS keyword_metrics (
    keyword_id      UUID NOT NULL,
    timestamp       TIMESTAMPTZ NOT NULL,
    volume          INT,
    cpc_cents       INT,
    competition     SMALLINT,
    heat_score      FLOAT,
    serp_position   SMALLINT
);

CREATE INDEX IF NOT EXISTS idx_kwm_kw_time
    ON keyword_metrics (keyword_id, timestamp DESC);

SELECT create_hypertable('keyword_metrics', 'timestamp', if_not_exists => TRUE);

-- ── Clusters ─────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS clusters (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    intent          VARCHAR(20),
    keyword_count   INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ── SERP Results (time series) ───────────────────────────────
CREATE TABLE IF NOT EXISTS serp_results (
    id          UUID DEFAULT gen_random_uuid(),
    keyword_id  UUID NOT NULL,
    position    SMALLINT NOT NULL,
    url         TEXT NOT NULL,
    title       TEXT,
    snippet     TEXT,
    crawled_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_serp_results_kw_crawled
    ON serp_results (keyword_id, crawled_at DESC);

SELECT create_hypertable('serp_results', 'crawled_at', if_not_exists => TRUE);

-- ── Page Data ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS page_data (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url                 TEXT NOT NULL UNIQUE,
    title               TEXT,
    meta_description    TEXT,
    h1                  TEXT[],
    h2                  TEXT[],
    h3                  TEXT[],
    word_count          INT,
    image_count         INT,
    video_count         INT,
    crawled_at          TIMESTAMPTZ DEFAULT NOW(),
    updated_at          TIMESTAMPTZ DEFAULT NOW()
);

-- ── Keyword-Page Relations (inverted index) ──────────────────
CREATE TABLE IF NOT EXISTS keyword_pages (
    keyword_id  UUID NOT NULL REFERENCES keywords(id) ON DELETE CASCADE,
    page_id     UUID NOT NULL REFERENCES page_data(id) ON DELETE CASCADE,
    position    SMALLINT,
    tfidf_score FLOAT,
    in_heading  BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (keyword_id, page_id)
);

CREATE INDEX IF NOT EXISTS idx_kp_keyword ON keyword_pages(keyword_id);
CREATE INDEX IF NOT EXISTS idx_kp_page ON keyword_pages(page_id);

-- ── Jobs (Asynq uses Redis, but we track metadata here) ──────
CREATE TABLE IF NOT EXISTS jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type        VARCHAR(50) NOT NULL,
    status      VARCHAR(20) DEFAULT 'PENDING',
    payload     JSONB,
    result      JSONB,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- ── Competitors ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS competitors (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    url         TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, name)
);
