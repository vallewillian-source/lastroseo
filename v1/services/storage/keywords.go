package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InsertKeyword inserts or updates (upserts) a keyword.
func InsertKeyword(ctx context.Context, pool *pgxpool.Pool, kw *Keyword) (*Keyword, error) {
	var k Keyword
	err := pool.QueryRow(ctx,
		`INSERT INTO keywords (project_id, keyword, is_seed, cluster_id, cluster_name, intent, source)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (project_id, keyword)
		 DO UPDATE SET is_seed=EXCLUDED.is_seed, intent=COALESCE(EXCLUDED.intent, keywords.intent),
		               source=EXCLUDED.source, updated_at=NOW()
		 RETURNING id, project_id, keyword, is_seed, cluster_id, cluster_name, intent, source, created_at, updated_at`,
		kw.ProjectID, kw.Keyword, kw.IsSeed, kw.ClusterID, kw.ClusterName, kw.Intent, kw.Source,
	).Scan(&k.ID, &k.ProjectID, &k.Keyword, &k.IsSeed, &k.ClusterID, &k.ClusterName,
		&k.Intent, &k.Source, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert keyword: %w", err)
	}
	return &k, nil
}

// GetKeyword returns a keyword by ID.
func GetKeyword(ctx context.Context, pool *pgxpool.Pool, id string) (*Keyword, error) {
	var k Keyword
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, keyword, is_seed, cluster_id, cluster_name, intent, source, created_at, updated_at
		 FROM keywords WHERE id=$1`, id,
	).Scan(&k.ID, &k.ProjectID, &k.Keyword, &k.IsSeed, &k.ClusterID, &k.ClusterName,
		&k.Intent, &k.Source, &k.CreatedAt, &k.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get keyword: %w", err)
	}
	return &k, nil
}

// ListKeywords returns paginated keywords for a project.
func ListKeywords(ctx context.Context, pool *pgxpool.Pool, projectID string, opts ListOpts) ([]Keyword, error) {
	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, keyword, is_seed, cluster_id, cluster_name, intent, source, created_at, updated_at
		 FROM keywords WHERE project_id=$1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		projectID, opts.Limit, opts.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list keywords: %w", err)
	}
	defer rows.Close()

	var keywords []Keyword
	for rows.Next() {
		var k Keyword
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.Keyword, &k.IsSeed, &k.ClusterID, &k.ClusterName,
			&k.Intent, &k.Source, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list keywords: scan: %w", err)
		}
		keywords = append(keywords, k)
	}
	return keywords, rows.Err()
}

// SearchKeywords uses PostgreSQL full-text search (Portuguese).
func SearchKeywords(ctx context.Context, pool *pgxpool.Pool, projectID, query string, limit int) ([]Keyword, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, keyword, is_seed, cluster_id, cluster_name, intent, source, created_at, updated_at
		 FROM keywords
		 WHERE project_id=$1 AND to_tsvector('portuguese', keyword) @@ plainto_tsquery('portuguese', $2)
		 ORDER BY ts_rank(to_tsvector('portuguese', keyword), plainto_tsquery('portuguese', $2)) DESC
		 LIMIT $3`,
		projectID, query, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search keywords: %w", err)
	}
	defer rows.Close()

	var keywords []Keyword
	for rows.Next() {
		var k Keyword
		if err := rows.Scan(&k.ID, &k.ProjectID, &k.Keyword, &k.IsSeed, &k.ClusterID, &k.ClusterName,
			&k.Intent, &k.Source, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, fmt.Errorf("search keywords: scan: %w", err)
		}
		keywords = append(keywords, k)
	}
	return keywords, rows.Err()
}

// UpdateKeywordCluster updates the cluster assignment for a keyword.
func UpdateKeywordCluster(ctx context.Context, pool *pgxpool.Pool, keywordID, clusterID, clusterName, intent string) error {
	_, err := pool.Exec(ctx,
		`UPDATE keywords SET cluster_id=$1, cluster_name=$2, intent=$3, updated_at=NOW()
		 WHERE id=$4`,
		clusterID, clusterName, intent, keywordID,
	)
	if err != nil {
		return fmt.Errorf("update keyword cluster: %w", err)
	}
	return nil
}

// BulkInsertKeywords inserts many keywords in a single batch (uses COPY internally).
func BulkInsertKeywords(ctx context.Context, pool *pgxpool.Pool, kws []Keyword) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("bulk insert keywords: begin: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, kw := range kws {
		_, err := tx.Exec(ctx,
			`INSERT INTO keywords (project_id, keyword, is_seed, source)
			 VALUES ($1,$2,$3,$4)
			 ON CONFLICT (project_id, keyword) DO UPDATE SET source=EXCLUDED.source, updated_at=NOW()`,
			kw.ProjectID, kw.Keyword, kw.IsSeed, kw.Source,
		)
		if err != nil {
			return fmt.Errorf("bulk insert keywords: %q: %w", kw.Keyword, err)
		}
	}
	return tx.Commit(ctx)
}
