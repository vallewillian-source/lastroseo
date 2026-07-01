-- 002: content gaps (analytics-svc LLM output)
CREATE TABLE IF NOT EXISTS content_gaps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    keyword_id UUID NOT NULL,
    keyword TEXT NOT NULL,
    gaps TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_content_gaps_project ON content_gaps (project_id, created_at DESC);
