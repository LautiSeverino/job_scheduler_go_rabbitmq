package repositories

import (
	"context"
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
	panic("unimplemented")
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
