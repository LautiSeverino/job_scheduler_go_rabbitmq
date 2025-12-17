package repositories

import (
	"context"
	"fmt"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"job_scheduler_go_rabbitmq/utils"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		qb.Query += fmt.Sprintf(" AND j.id = $%d", len(qb.Args)+1)
		qb.Args = append(qb.Args, params.ID)
	}
	if params.Status != nil {
		qb.Query += fmt.Sprintf(" AND j.status = $%d", len(qb.Args)+1)
		qb.Args = append(qb.Args, params.Status)
	}
	if params.Type != nil {
		qb.Query += fmt.Sprintf(" AND j.type = $%d", len(qb.Args)+1)
		qb.Args = append(qb.Args, params.Type)
	}

	if params.ReadyToRun != nil && *params.ReadyToRun {
		now := time.Now()

		qb.Query += fmt.Sprintf(
			" AND (j.scheduled_at IS NULL OR j.scheduled_at <= $%d)",
			len(qb.Args)+1,
		)
		qb.Args = append(qb.Args, now)
	}

	if params.LockFree != nil && *params.LockFree {
		if params.LockTimeout != nil {
			expiredAt := time.Now().Add(-*params.LockTimeout)

			qb.Query += fmt.Sprintf(
				" AND (j.locked_at IS NULL OR j.locked_at < $%d)",
				len(qb.Args)+1,
			)
			qb.Args = append(qb.Args, expiredAt)
		} else {
			qb.Query += " AND j.locked_at IS NULL"
		}
	}

	return nil
}

// Insert implements ports.IJobRepository.
func (r *JobRepository) Insert(ctx context.Context, job domain.Job) error {
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

	var err error

	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}

	if err != nil {
		return fmt.Errorf("failed to execute insert: %w", err)
	}

	return nil
}

// LockJob implements ports.IJobRepository.
func (r *JobRepository) LockJob(ctx context.Context, jobID uuid.UUID, lockedBy string) error {
	now := time.Now()
	lockExpiry := now.Add(-5 * time.Minute)

	query := utils.QueryBuilder{
		Query: `
			UPDATE jobs
			SET
				locked_at = $1,
				locked_by = $2,
				updated_at = $1
			WHERE id = $3
			AND status = $4
			AND (
				locked_at IS NULL
				OR locked_at < $5
			)
		`,
		Args: []any{
			now,
			lockedBy,
			jobID,
			domain.JobStatusPending,
			lockExpiry,
		},
	}

	var cmdTag pgconn.CommandTag
	var err error

	if r.tx != nil {
		cmdTag, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		cmdTag, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}

	if err != nil {
		return fmt.Errorf("lock job failed: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("job already locked or not pending")
	}

	return nil
}

// MarkCompleted implements ports.IJobRepository.
func (r *JobRepository) MarkCompleted(ctx context.Context, jobID uuid.UUID) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE jobs
		SET
			status = $1,
			completed_at = $2,
			locked_at = NULL,
			locked_by = NULL,
			updated_at = $2
		WHERE id = $3
		AND status = $4
	`,
		Args: []any{domain.JobStatusCompleted, time.Now(), jobID, domain.JobStatusRunning},
	}
	var err error

	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}

	if err != nil {
		return fmt.Errorf("mark completed failed: %w", err)
	}
	return nil
}

// MarkFailed implements ports.IJobRepository.
func (r *JobRepository) MarkFailed(ctx context.Context, jobID uuid.UUID, errMsg string, httpStatus *int) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE jobs
		SET
			status = $1,
			locked_at = NULL,
			locked_by = NULL,
			updated_at = $2
		WHERE id = $3
		AND status = $4
	`,
		Args: []any{domain.JobStatusFailed, time.Now(), jobID, domain.JobStatusRunning},
	}
	var err error

	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}

	if err != nil {
		return fmt.Errorf("mark failed failed: %w", err)
	}
	return nil
}

// MarkQueued implements ports.IJobRepository.
func (r *JobRepository) MarkQueued(ctx context.Context, jobID uuid.UUID) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE jobs
		SET
			status = $1,
			updated_at = $2
		WHERE id = $3
		AND status = $4
	`,
		Args: []any{domain.JobStatusQueued, time.Now(), jobID, domain.JobStatusPending},
	}

	var err error

	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}

	if err != nil {
		return fmt.Errorf("mark queued failed: %w", err)
	}
	return nil
}

// MarkRunning implements ports.IJobRepository.
func (r *JobRepository) MarkRunning(ctx context.Context, jobID uuid.UUID) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE jobs
		SET
			status = $1,
			updated_at = $2
		WHERE id = $3
		AND status = $4
	`,
		Args: []any{domain.JobStatusRunning, time.Now(), jobID, domain.JobStatusQueued},
	}
	var err error

	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}
	if err != nil {
		return fmt.Errorf("mark running failed: %w", err)
	}
	return nil
}

// MarkDead implements ports.IJobRepository.
func (r *JobRepository) MarkDead(ctx context.Context, jobID uuid.UUID, reason string) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE jobs
		SET
			status = $1,
			locked_at = NULL,
			locked_by = NULL,
			updated_at = $2
		WHERE id = $3
		AND status = $4
	`,
		Args: []any{domain.JobStatusDead, time.Now(), jobID, domain.JobStatusFailed},
	}

	var err error

	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}
	if err != nil {
		return fmt.Errorf("mark dead failed: %w", err)
	}

	return nil
}
