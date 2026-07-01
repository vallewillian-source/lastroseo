package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateJob inserts a new job record.
func CreateJob(ctx context.Context, pool *pgxpool.Pool, job *Job) (*Job, error) {
	var j Job
	err := pool.QueryRow(ctx,
		`INSERT INTO jobs (project_id, type, status, payload)
		 VALUES ($1,$2,$3,$4)
		 RETURNING id, project_id, type, status, created_at, updated_at`,
		job.ProjectID, job.Type, job.Status, job.Payload,
	).Scan(&j.ID, &j.ProjectID, &j.Type, &j.Status, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create job: %w", err)
	}
	return &j, nil
}

// GetJob returns a job by ID.
func GetJob(ctx context.Context, pool *pgxpool.Pool, id string) (*Job, error) {
	var j Job
	err := pool.QueryRow(ctx,
		`SELECT id, project_id, type, status, payload, result, created_at, updated_at
		 FROM jobs WHERE id=$1`, id,
	).Scan(&j.ID, &j.ProjectID, &j.Type, &j.Status, &j.Payload, &j.Result, &j.CreatedAt, &j.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get job: %w", err)
	}
	return &j, nil
}

// UpdateJobStatus updates the status and optionally the result of a job.
func UpdateJobStatus(ctx context.Context, pool *pgxpool.Pool, id string, status JobStatus, result []byte) error {
	_, err := pool.Exec(ctx,
		`UPDATE jobs SET status=$1, result=$2, updated_at=NOW() WHERE id=$3`,
		status, result, id,
	)
	if err != nil {
		return fmt.Errorf("update job status: %w", err)
	}
	return nil
}

// ListJobsByProject returns jobs for a project, newest first.
func ListJobsByProject(ctx context.Context, pool *pgxpool.Pool, projectID string, opts ListOpts) ([]Job, error) {
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	rows, err := pool.Query(ctx,
		`SELECT id, project_id, type, status, payload, result, created_at, updated_at
		 FROM jobs WHERE project_id=$1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		projectID, opts.Limit, opts.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.ProjectID, &j.Type, &j.Status, &j.Payload, &j.Result, &j.CreatedAt, &j.UpdatedAt); err != nil {
			return nil, fmt.Errorf("list jobs: scan: %w", err)
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}
