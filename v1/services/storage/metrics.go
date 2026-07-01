package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InsertMetric inserts a single keyword metric data point.
func InsertMetric(ctx context.Context, pool *pgxpool.Pool, m *KeywordMetric) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO keyword_metrics (keyword_id, timestamp, volume, cpc_cents, competition, heat_score, serp_position)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (keyword_id, timestamp) DO UPDATE SET
		   volume=EXCLUDED.volume, cpc_cents=EXCLUDED.cpc_cents,
		   competition=EXCLUDED.competition, heat_score=EXCLUDED.heat_score,
		   serp_position=EXCLUDED.serp_position`,
		m.KeywordID, m.Timestamp, m.Volume, m.CPCCents, m.Competition, m.HeatScore, m.SerpPosition,
	)
	if err != nil {
		return fmt.Errorf("insert metric: %w", err)
	}
	return nil
}

// GetLatestMetrics returns the most recent metric for each keyword in a project.
func GetLatestMetrics(ctx context.Context, pool *pgxpool.Pool, projectID string) ([]KeywordMetric, error) {
	rows, err := pool.Query(ctx,
		`SELECT DISTINCT ON (km.keyword_id)
		   km.keyword_id, km.timestamp, km.volume, km.cpc_cents, km.competition, km.heat_score, km.serp_position
		 FROM keyword_metrics km
		 JOIN keywords k ON k.id = km.keyword_id
		 WHERE k.project_id=$1
		 ORDER BY km.keyword_id, km.timestamp DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("get latest metrics: %w", err)
	}
	defer rows.Close()

	var metrics []KeywordMetric
	for rows.Next() {
		var m KeywordMetric
		if err := rows.Scan(&m.KeywordID, &m.Timestamp, &m.Volume, &m.CPCCents, &m.Competition, &m.HeatScore, &m.SerpPosition); err != nil {
			return nil, fmt.Errorf("get latest metrics: scan: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// GetMetricHistory returns time-series data for a keyword.
func GetMetricHistory(ctx context.Context, pool *pgxpool.Pool, keywordID string, since time.Time, limit int) ([]KeywordMetric, error) {
	if limit <= 0 {
		limit = 52 // 52 weeks
	}
	rows, err := pool.Query(ctx,
		`SELECT keyword_id, timestamp, volume, cpc_cents, competition, heat_score, serp_position
		 FROM keyword_metrics
		 WHERE keyword_id=$1 AND timestamp >= $2
		 ORDER BY timestamp DESC LIMIT $3`,
		keywordID, since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get metric history: %w", err)
	}
	defer rows.Close()

	var metrics []KeywordMetric
	for rows.Next() {
		var m KeywordMetric
		if err := rows.Scan(&m.KeywordID, &m.Timestamp, &m.Volume, &m.CPCCents, &m.Competition, &m.HeatScore, &m.SerpPosition); err != nil {
			return nil, fmt.Errorf("get metric history: scan: %w", err)
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// ── SERP Results ──────────────────────────────────────────────

// InsertSERPResult inserts a single SERP result.
func InsertSERPResult(ctx context.Context, pool *pgxpool.Pool, r *SERPResult) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO serp_results (keyword_id, position, url, title, snippet, crawled_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		r.KeywordID, r.Position, r.URL, r.Title, r.Snippet, r.CrawledAt,
	)
	if err != nil {
		return fmt.Errorf("insert serp result: %w", err)
	}
	return nil
}

// BulkInsertSERPResults inserts many SERP results in a single batch.
func BulkInsertSERPResults(ctx context.Context, pool *pgxpool.Pool, results []SERPResult) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("bulk insert serp: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for i := range results {
		_, err := tx.Exec(ctx,
			`INSERT INTO serp_results (keyword_id, position, url, title, snippet, crawled_at)
			 VALUES ($1,$2,$3,$4,$5,$6)`,
			results[i].KeywordID, results[i].Position, results[i].URL,
			results[i].Title, results[i].Snippet, results[i].CrawledAt,
		)
		if err != nil {
			return fmt.Errorf("bulk insert serp: row %d: %w", i, err)
		}
	}
	return tx.Commit(ctx)
}

// SERPProjectResult joins SERPResult with its keyword text.
type SERPProjectResult struct {
	SERPResult
	Keyword string `json:"keyword"`
}

// ListSERPResultsByProject returns all SERP results for a project joined with keyword text.
func ListSERPResultsByProject(ctx context.Context, pool *pgxpool.Pool, projectID string, opts ListOpts) ([]SERPProjectResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 200
	}
	rows, err := pool.Query(ctx,
		`SELECT sr.id, sr.keyword_id, sr.position, sr.url, sr.title, sr.snippet, sr.crawled_at,
		        k.keyword
		 FROM serp_results sr
		 JOIN keywords k ON k.id = sr.keyword_id
		 WHERE k.project_id = $1
		 ORDER BY k.keyword, sr.position
		 LIMIT $2 OFFSET $3`,
		projectID, opts.Limit, opts.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list serp results by project: %w", err)
	}
	defer rows.Close()

	var results []SERPProjectResult
	for rows.Next() {
		var r SERPProjectResult
		if err := rows.Scan(&r.ID, &r.KeywordID, &r.Position, &r.URL, &r.Title, &r.Snippet, &r.CrawledAt,
			&r.Keyword); err != nil {
			return nil, fmt.Errorf("list serp results by project: scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetSERPHistory returns SERP snapshots for a keyword.
func GetSERPHistory(ctx context.Context, pool *pgxpool.Pool, keywordID string, since time.Time, limit int) ([]SERPResult, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := pool.Query(ctx,
		`SELECT id, keyword_id, position, url, title, snippet, crawled_at
		 FROM serp_results
		 WHERE keyword_id=$1 AND crawled_at >= $2
		 ORDER BY crawled_at DESC LIMIT $3`,
		keywordID, since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get serp history: %w", err)
	}
	defer rows.Close()

	var results []SERPResult
	for rows.Next() {
		var r SERPResult
		if err := rows.Scan(&r.ID, &r.KeywordID, &r.Position, &r.URL, &r.Title, &r.Snippet, &r.CrawledAt); err != nil {
			return nil, fmt.Errorf("get serp history: scan: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
