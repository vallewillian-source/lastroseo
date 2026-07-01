package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetContentGaps returns all content gaps for a project.
func GetContentGaps(ctx context.Context, pool *pgxpool.Pool, projectID string, opts ListOpts) ([]ContentGap, error) {
	if opts.Limit <= 0 {
		opts.Limit = 50
	}
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, keyword_id, keyword, gaps, created_at
		 FROM content_gaps WHERE project_id=$1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		projectID, opts.Limit, opts.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("get content gaps: %w", err)
	}
	defer rows.Close()

	var gaps []ContentGap
	for rows.Next() {
		var g ContentGap
		if err := rows.Scan(&g.ID, &g.ProjectID, &g.KeywordID, &g.Keyword, &g.Gaps, &g.CreatedAt); err != nil {
			return nil, fmt.Errorf("get content gaps: scan: %w", err)
		}
		gaps = append(gaps, g)
	}
	return gaps, rows.Err()
}
