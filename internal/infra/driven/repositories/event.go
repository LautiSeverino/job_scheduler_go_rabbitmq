package repositories

import (
	"context"
	"fmt"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"job_scheduler_go_rabbitmq/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepository struct {
	tx   pgx.Tx
	pool *pgxpool.Pool
}

func NewEventRepository(tx pgx.Tx, pool *pgxpool.Pool) ports.IEventRepository {
	return &EventRepository{tx: tx, pool: pool}
}

// Get implements ports.IEventRepository.
func (r *EventRepository) Get(ctx context.Context, params domain.EventSearchParams) ([]domain.Event, error) {
	query := utils.QueryBuilder{
		Query: `
		SELECT
			id,
			job_id,
			type,
			message,
			metadata,
			created_at
		FROM events
		WHERE 1=1
	`,
	}

	if err := r.buildSearchParams(&query, params); err != nil {
		return nil, fmt.Errorf("build search params: %w", err)
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

	var events []domain.Event
	for rows.Next() {
		var event domain.Event
		if err := rows.Scan(
			&event.ID,
			&event.JobID,
			&event.Type,
			&event.Message,
			&event.Metadata,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *EventRepository) buildSearchParams(query *utils.QueryBuilder, params domain.EventSearchParams) error {
	if params.ID != nil {
		query.Query += fmt.Sprintf(" AND id = $%d", len(query.Args)+1)
		query.Args = append(query.Args, *params.ID)
	}
	if params.JobID != nil {
		query.Query += fmt.Sprintf(" AND job_id = $%d", len(query.Args)+1)
		query.Args = append(query.Args, *params.JobID)
	}

	return nil
}

// Insert implements ports.IEventRepository.
func (r *EventRepository) Insert(ctx context.Context, event domain.Event) error {
	query := utils.QueryBuilder{
		Query: `
		INSERT INTO events (
			id,
			job_id,
			type,
			message,
			metadata,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`,
		Args: []any{
			event.ID,
			event.JobID,
			event.Type,
			event.Message,
			event.Metadata,
			event.CreatedAt,
		},
	}

	var err error
	if r.tx != nil {
		_, err = r.tx.Exec(ctx, query.Query, query.Args...)

	} else {
		_, err = r.pool.Exec(ctx, query.Query, query.Args...)
	}
	if err != nil {
		return err
	}
	return nil
}
