package repositories

import (
	"context"
	"fmt"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"job_scheduler_go_rabbitmq/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepository struct {
	tx   pgx.Tx
	pool *pgxpool.Pool
}

func NewJobRepository(tx pgx.Tx, pool *pgxpool.Pool) ports.IJobRepository {
	return &JobRepository{tx: tx, pool: pool}
}

// GetDueJobs implements ports.IJobRepository.
func (r *JobRepository) Get(ctx context.Context, params domain.JobSearchParams) ([]domain.Job, error) {
	query := utils.QueryBuilder{
		Query: ` SELECT
				j.id,
				j.type,
				j.callback_url,
				j.payload,
				j.status,
				j.max_retries,
				j.scheduled_at,
				j.locked_at,
				j.locked_by,
				j.completed_at,
				j.priority,
				j.created_at,
				j.updated_at
			FROM jobs j
			WHERE 1=1
		`,
		Args: []any{},
	}

	if err := r.buildSearchParams(&query, params); err != nil {
		return nil, fmt.Errorf("failed to build search params: %w", err)
	}

	var rows pgx.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.Query(ctx, query.Query, query.Args...)
	} else {
		rows, err = r.pool.Query(ctx, query.Query, query.Args...)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var job domain.Job
		err := rows.Scan(
			&job.ID,
			&job.Type,
			&job.CallbackURL,
			&job.Payload,
			&job.Status,
			&job.MaxRetries,
			&job.ScheduledAt,
			&job.LockedAt,
			&job.LockedBy,
			&job.CompletedAt,
			&job.Priority,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return jobs, nil

}

// GetOne implements ports.IJobRepository.
func (r *JobRepository) GetOne(ctx context.Context, params domain.JobSearchParams) (*domain.Job, error) {
	query := utils.QueryBuilder{
		Query: ` SELECT
				j.id,
				j.type,
				j.callback_url,
				j.payload,
				j.status,
				j.max_retries,
				j.scheduled_at,
				j.locked_at,
				j.locked_by,
				j.completed_at,
				j.priority,
				j.created_at,
				j.updated_at
			FROM jobs j
			WHERE 1=1
		`,
		Args: []any{},
	}

	if err := r.buildSearchParams(&query, params); err != nil {
		return nil, fmt.Errorf("failed to build search params: %w", err)
	}

	var row pgx.Row
	if r.tx != nil {
		row = r.tx.QueryRow(ctx, query.Query, query.Args...)
	} else {
		row = r.pool.QueryRow(ctx, query.Query, query.Args...)
	}

	var job domain.Job
	err := row.Scan(
		&job.ID,
		&job.Type,
		&job.CallbackURL,
		&job.Payload,
		&job.Status,
		&job.MaxRetries,
		&job.ScheduledAt,
		&job.LockedAt,
		&job.LockedBy,
		&job.CompletedAt,
		&job.Priority,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("[TEAM][REPOSITORY][GetOne()] Error en Scan: %v", err)
	}

	return &job, nil
}

func (r *JobRepository) buildSearchParams(qb *utils.QueryBuilder, params domain.JobSearchParams) error {
	if params.ID != nil {
		qb.Query += fmt.Sprintf(" AND  = $%d", len(qb.Args)+1)
		qb.Args = append(qb.Args, params.ID)
	}
	if params.Status != nil {
		qb.Query += fmt.Sprintf(" AND status = $%d", len(qb.Args)+1)
		qb.Args = append(qb.Args, params.Status)
	}
	if params.Type != nil {
		qb.Query += fmt.Sprintf(" AND type = $%d", len(qb.Args)+1)
		qb.Args = append(qb.Args, params.Type)
	}

	return nil
}

// Insert implements ports.IJobRepository.
func (r *JobRepository) Insert(ctx context.Context, job domain.Job) (*domain.Job, error) {
	query := utils.QueryBuilder{
		Query: `
		INSERT INTO jobs
		(id, 
		type, 
		callback_url, 
		payload, 
		status, 
		max_retries, 
		scheduled_at,
		priority, 
		created_at, 
		updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		Args: []any{
			job.ID,
			job.Type,
			job.CallbackURL,
			job.Payload,
			job.Status,
			job.MaxRetries,
			job.ScheduledAt,
			job.Priority,
			job.CreatedAt,
			job.UpdatedAt,
		},
	}

	_, err := r.tx.Exec(ctx, query.Query, query.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute insert: %w", err)
	}

	return &job, nil

}

// LockJob implements ports.IJobRepository.
func (r *JobRepository) LockJob(ctx context.Context, jobID uuid.UUID, workerID string, status domain.JobStatus) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE jobs
		SET locked_at = NOW(),
		locked_by = $1,
		status = $2,
		updated_at = NOW()
		WHERE id = $3 AND (locked_at IS NULL OR locked_at < NOW() - INTERVAL '5 minutes')
		`,
		Args: []any{
			workerID,
			status,
			jobID,
		},
	}

	cmdTag, err := r.tx.Exec(ctx, query.Query, query.Args...)
	if err != nil {
		return fmt.Errorf("failed to execute lock job: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("job is already locked or does not exist")
	}

	return nil
}

// Update implements ports.IJobRepository.
func (r *JobRepository) Update(ctx context.Context, job domain.Job) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE jobs
		SET type = $1,
		callback_url = $2,
		payload = $3,
		status = $4,
		max_retries = $5,
		scheduled_at = $6,
		locked_at = $7,
		locked_by = $8,
		completed_at = $9,
		priority = $10,
		updated_at = $11
		WHERE id = $12
		`,
		Args: []any{
			job.Type,
			job.CallbackURL,
			job.Payload,
			job.Status,
			job.MaxRetries,
			job.ScheduledAt,
			job.LockedAt,
			job.LockedBy,
			job.CompletedAt,
			job.Priority,
			job.UpdatedAt,
			job.ID,
		},
	}

	_, err := r.tx.Exec(ctx, query.Query, query.Args...)
	if err != nil {
		return fmt.Errorf("failed to execute update: %w", err)
	}

	return nil
}
