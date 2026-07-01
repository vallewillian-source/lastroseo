package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateCompetitor inserts a new competitor for a project.
func CreateCompetitor(ctx context.Context, pool *pgxpool.Pool, projectID, name, url string) (*Competitor, error) {
	var c Competitor
	err := pool.QueryRow(ctx,
		`INSERT INTO competitors (project_id, name, url) VALUES ($1,$2,$3)
		 RETURNING id, project_id, name, url, created_at`,
		projectID, name, url,
	).Scan(&c.ID, &c.ProjectID, &c.Name, &c.URL, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create competitor: %w", err)
	}
	return &c, nil
}

// ListCompetitors returns all competitors for a project.
func ListCompetitors(ctx context.Context, pool *pgxpool.Pool, projectID string) ([]Competitor, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, name, url, created_at
		 FROM competitors WHERE project_id=$1 ORDER BY created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list competitors: %w", err)
	}
	defer rows.Close()

	var competitors []Competitor
	for rows.Next() {
		var c Competitor
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Name, &c.URL, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("list competitors: scan: %w", err)
		}
		competitors = append(competitors, c)
	}
	return competitors, rows.Err()
}

// GetCompetitor returns a competitor by ID.
func GetCompetitor(ctx context.Context, pool *pgxpool.Pool, id string) (*Competitor, error) {
	var c Competitor
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, name, url, created_at
		 FROM competitors WHERE id=$1`, id,
	).Scan(&c.ID, &c.ProjectID, &c.Name, &c.URL, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get competitor: %w", err)
	}
	return &c, nil
}

// DeleteCompetitor removes a competitor by ID.
func DeleteCompetitor(ctx context.Context, pool *pgxpool.Pool, id string) error {
	tag, err := pool.Exec(ctx,
		`DELETE FROM competitors WHERE id=$1`, id,
	)
	if err != nil {
		return fmt.Errorf("delete competitor: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete competitor: not found")
	}
	return nil
}
