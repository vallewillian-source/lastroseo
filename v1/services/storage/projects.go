package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateProject inserts a new project and its seed keywords in a transaction.
func CreateProject(ctx context.Context, pool *pgxpool.Pool, name, businessDesc, targetAudience string, seedKeywords []string) (*Project, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("create project: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var p Project
	err = tx.QueryRow(ctx,
		`INSERT INTO projects (name, business_desc, target_audience) VALUES ($1,$2,$3)
		 RETURNING id, name, business_desc, target_audience, created_at, updated_at`,
		name, businessDesc, targetAudience,
	).Scan(&p.ID, &p.Name, &p.BusinessDesc, &p.TargetAudience, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create project: insert: %w", err)
	}

	for _, kw := range seedKeywords {
		_, err := tx.Exec(ctx,
			`INSERT INTO seed_keywords (project_id, keyword) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			p.ID, kw,
		)
		if err != nil {
			return nil, fmt.Errorf("create project: seed keyword %q: %w", kw, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("create project: commit: %w", err)
	}
	return &p, nil
}

// GetProject returns a project by ID.
func GetProject(ctx context.Context, pool *pgxpool.Pool, id string) (*Project, error) {
	var p Project
	err := pool.QueryRow(ctx,
		`SELECT id, name, business_desc, target_audience, created_at, updated_at
		 FROM projects WHERE id=$1`, id,
	).Scan(&p.ID, &p.Name, &p.BusinessDesc, &p.TargetAudience, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get project: %w", err)
	}
	return &p, nil
}

// ListProjects returns projects with pagination.
func ListProjects(ctx context.Context, pool *pgxpool.Pool, opts ListOpts) ([]Project, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	rows, err := pool.Query(ctx,
		`SELECT id, name, business_desc, target_audience, created_at, updated_at
		 FROM projects ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		opts.Limit, opts.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.BusinessDesc, &p.TargetAudience, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list projects: scan: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// AddSeedKeywords inserts seed keywords for a project (idempotent).
func AddSeedKeywords(ctx context.Context, pool *pgxpool.Pool, projectID string, keywords []string) error {
	for _, kw := range keywords {
		_, err := pool.Exec(ctx,
			`INSERT INTO seed_keywords (project_id, keyword) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
			projectID, kw,
		)
		if err != nil {
			return fmt.Errorf("add seed keyword %q: %w", kw, err)
		}
	}
	return nil
}
