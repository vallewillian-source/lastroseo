package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateCluster inserts a new keyword cluster.
func CreateCluster(ctx context.Context, pool *pgxpool.Pool, c *Cluster) (*Cluster, error) {
	var out Cluster
	err := pool.QueryRow(ctx,
		`INSERT INTO clusters (project_id, name, intent, keyword_count)
		 VALUES ($1,$2,$3,$4)
		 RETURNING id, project_id, name, intent, keyword_count, created_at, updated_at`,
		c.ProjectID, c.Name, c.Intent, c.KeywordCount,
	).Scan(&out.ID, &out.ProjectID, &out.Name, &out.Intent, &out.KeywordCount, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create cluster: %w", err)
	}
	return &out, nil
}

// GetClustersByProject returns all clusters for a project.
func GetClustersByProject(ctx context.Context, pool *pgxpool.Pool, projectID string) ([]Cluster, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, name, intent, keyword_count, created_at, updated_at
		 FROM clusters WHERE project_id=$1 ORDER BY keyword_count DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("get clusters: %w", err)
	}
	defer rows.Close()

	var clusters []Cluster
	for rows.Next() {
		var c Cluster
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.Intent, &c.KeywordCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("get clusters: scan: %w", err)
		}
		clusters = append(clusters, c)
	}
	return clusters, rows.Err()
}

// UpdateClusterKeywordCount updates the keyword count for a cluster.
func UpdateClusterKeywordCount(ctx context.Context, pool *pgxpool.Pool, clusterID string, count int) error {
	_, err := pool.Exec(ctx,
		`UPDATE clusters SET keyword_count=$1, updated_at=NOW() WHERE id=$2`,
		count, clusterID,
	)
	if err != nil {
		return fmt.Errorf("update cluster count: %w", err)
	}
	return nil
}
