package ports

import (
	"context"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type IJobHandler interface {
	RegisterRouter(router *mux.Router)
	Create() http.HandlerFunc
	GetOne() http.HandlerFunc
	GetTimeline() http.HandlerFunc
}

type IJobService interface {
	Create(ctx context.Context, input domain.CreateJobInput) (*domain.Job, error)
	GetOne(ctx context.Context, params domain.JobSearchParams) (*domain.Job, error)
	GetTimeline(ctx context.Context, jobID uuid.UUID) ([]domain.Event, error)
}

type IJobExecutionService interface {
	ProcessJobMessage(ctx context.Context, msg domain.RabbitJobMessage) error
}

type IJobRepository interface {
	// Base
	Insert(ctx context.Context, job domain.Job) error
	GetOne(ctx context.Context, params domain.JobSearchParams) (*domain.Job, error)
	Get(ctx context.Context, params domain.JobSearchParams) ([]domain.Job, error)

	// Dispatcher
	LockJob(ctx context.Context, jobID uuid.UUID, lockedBy string) error
	MarkQueued(ctx context.Context, jobID uuid.UUID) error

	// Worker
	MarkRunning(ctx context.Context, jobID uuid.UUID) error
	MarkCompleted(ctx context.Context, jobID uuid.UUID) error
	MarkFailed(ctx context.Context, jobID uuid.UUID, errMsg string, httpStatus *int) error
	MarkDead(ctx context.Context, jobID uuid.UUID, reason string) error
}

// intento de ejecutar un job
type IAttemptRepository interface {
	Insert(ctx context.Context, attempt domain.Attempt) error
	MarkSuccess(ctx context.Context, attemptID uuid.UUID) error
	MarkFailed(ctx context.Context, attemptID uuid.UUID, errMsg string, httpStatus *int) error
	Get(ctx context.Context, params domain.AttemptSearchParams) ([]domain.Attempt, error)
	Count(ctx context.Context, params domain.AttemptSearchParams) (int, error)
}

type IEventRepository interface {
	Insert(ctx context.Context, event domain.Event) error
	Get(ctx context.Context, params domain.EventSearchParams) ([]domain.Event, error)
}

type IJobExecutor interface {
	Execute(ctx context.Context, job *domain.Job) domain.ExecutionResult
}
