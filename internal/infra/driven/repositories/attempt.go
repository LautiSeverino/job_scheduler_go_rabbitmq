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
	"github.com/jackc/pgx/v5/pgxpool"
)

type AttemptRepository struct {
	tx   pgx.Tx
	pool *pgxpool.Pool
}

func NewAttemptRepository(tx pgx.Tx, pool *pgxpool.Pool) ports.IAttemptRepository {
	return &AttemptRepository{tx: tx, pool: pool}
}

// MarkFailed implements ports.IAttemptRepository.
func (r *AttemptRepository) MarkFailed(ctx context.Context, attemptID uuid.UUID, errMsg string, httpStatus *int) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE job_attempts
		SET
			status = $1,
			error_message = $2,
			finished_at = $3,
			http_status = $4
		WHERE id = $5
	`,
		Args: []any{domain.AttemptStatusFailed, errMsg, time.Now(), httpStatus, attemptID},
	}

	_, err := r.tx.Exec(ctx, query.Query, query.Args...)
	if err != nil {
		return fmt.Errorf("mark attempt failed: %w", err)
	}

	return nil
}

// MarkSuccess implements ports.IAttemptRepository.
func (r *AttemptRepository) MarkSuccess(ctx context.Context, attemptID uuid.UUID) error {
	query := utils.QueryBuilder{
		Query: `
		UPDATE job_attempts
		SET
			status = $1,
			finished_at = $2
		WHERE id = $3
	`,
		Args: []any{domain.AttemptStatusSuccess, time.Now(), attemptID},
	}

	_, err := r.tx.Exec(ctx, query.Query, query.Args...)
	if err != nil {
		return fmt.Errorf("mark attempt success failed: %w", err)
	}

	return nil
}

// Count implements ports.IJobAttemptRepository.
func (r *AttemptRepository) Count(ctx context.Context, params domain.AttemptSearchParams) (int, error) {
	query := &utils.QueryBuilder{
		Query: `
		SELECT COUNT (DISTINCT a.id)
			FROM job_attempts a
			WHERE 1=1`,
		Args: []any{},
	}

	r.buildSearchParams(query, params)

	var count int
	var row pgx.Row

	if r.tx != nil {
		row = r.tx.QueryRow(ctx, query.Query, query.Args...)
	} else {
		row = r.pool.QueryRow(ctx, query.Query, query.Args...)
	}

	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return count, nil
}

// Create implements ports.IJobAttemptRepository.
func (r *AttemptRepository) Insert(ctx context.Context, attempt domain.Attempt) error {
	query := utils.QueryBuilder{
		Query: ` INSERT INTO job_attempts
				(
				id,
				job_id,
				attempt_number,
				started_at,
				status,
				error_message,
				http_status,
				created_at
				)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		Args: []any{
			attempt.ID,
			attempt.JobID,
			attempt.AttemptNumber,
			attempt.StartedAt,
			attempt.Status,
			attempt.ErrorMessage,
			attempt.HTTPStatus,
			attempt.CreatedAt,
		},
	}

	var err error
	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)
	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}

	if err != nil {
		return fmt.Errorf("error al insertar el attempt: %w", err)
	}

	return nil
}

// Get implements ports.IJobAttemptRepository.
func (r *AttemptRepository) Get(ctx context.Context, params domain.AttemptSearchParams) ([]domain.Attempt, error) {
	query := utils.QueryBuilder{
		Query: ` SELECT
				a.id,
				a.job_id,
				a.attempt_number,
				a.started_at,
				a.status,
				a.error_message,
				a.http_status,
				a.created_at
			FROM job_attempts a
			WHERE 1=1
		`,
		Args: []any{},
	}

	if err := r.buildSearchParams(&query, params); err != nil {
		return nil, err
	}

	var rows pgx.Rows
	var err error
	if r.tx != nil {
		rows, err = r.tx.Query(ctx, query.Query, query.Args...)
	} else {
		rows, err = r.pool.Query(ctx, query.Query, query.Args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []domain.Attempt
	for rows.Next() {
		var attempt domain.Attempt
		if err := rows.Scan(
			&attempt.ID,
			&attempt.JobID,
			&attempt.AttemptNumber,
			&attempt.StartedAt,
			&attempt.Status,
			&attempt.ErrorMessage,
			&attempt.HTTPStatus,
			&attempt.CreatedAt,
		); err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	return attempts, nil
}

func (r *AttemptRepository) buildSearchParams(query *utils.QueryBuilder, params domain.AttemptSearchParams) error {
	if params.ID != nil {
		query.Query += fmt.Sprintf(" AND a.id = $%d", len(query.Args)+1)
		query.Args = append(query.Args, *params.ID)

	}
	if params.JobID != nil {
		query.Query += fmt.Sprintf(" AND a.job_id = $%d", len(query.Args)+1)
		query.Args = append(query.Args, *params.JobID)
	}
	return nil
}
