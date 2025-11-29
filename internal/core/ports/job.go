package ports

import (
	"context"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type IJobService interface {

	//Registrar trabajos
	Create(ctx context.Context, input domain.CreateJobInput) (*domain.Job, error)

	//Detectar jobs vencidos y encolarlos
	EnqueueDueJobs(ctx context.Context) error

	//Tomar un job para ejecutar
	LockNextJob(ctx context.Context, workerID string) (*domain.Job, error)

	//Ejecutar ciclo de vida
	MarkJobRunning(ctx context.Context, jobID uuid.UUID) error
	MarkJobSuccess(ctx context.Context, jobID uuid.UUID) error
	MarkJobFailed(ctx context.Context, jobID uuid.UUID, errMsg string, httpStatus *int) error

	// Reintentos
	RetryJob(ctx context.Context, jobID uuid.UUID) error

	// Dead Letter
	MoveToDeadLetter(ctx context.Context, jobID uuid.UUID, reason string) error

	GetOne(ctx context.Context, params domain.JobSearchParams) (*domain.Job, error)
	GetJobTimeline(ctx context.Context, jobID uuid.UUID) ([]domain.JobEvent, error)
}

type IJobRepository interface {
	Insert(ctx context.Context, job domain.Job) (*domain.Job, error)
	Update(ctx context.Context, job domain.Job) error
	GetOne(ctx context.Context, params domain.JobSearchParams) (*domain.Job, error)
	Get(ctx context.Context, params domain.JobSearchParams) ([]domain.Job, error)
	LockJob(ctx context.Context, jobID uuid.UUID, workerID string, status domain.JobStatus) error
}

type IJobAttemptRepository interface {
	Create(ctx context.Context, attempt domain.JobAttempt) error
	UpdateStatus(ctx context.Context, attemptID uint64, status domain.AttemptStatus, errMsg *string, httpStatus *int) error

	Get(ctx context.Context, params domain.JobAttemptSearchParams) ([]domain.JobAttempt, error)
	Count(ctx context.Context, params domain.JobAttemptSearchParams) (int, error)
}

type IJobEventRepository interface {
	Insert(ctx context.Context, event domain.JobEvent) error
	Get(ctx context.Context, params domain.JobEventSearchParams) ([]domain.JobEvent, error)
}

type IDeadLetterRepository interface {
	Insert(ctx context.Context, dead domain.JobDeadLetter) error
	Get(ctx context.Context, params domain.JobDeadLetterSearchParams) ([]domain.JobDeadLetter, error)
	Delete(ctx context.Context, jobID uuid.UUID) error
}

type IJobHandler interface {
	Create() http.HandlerFunc
	EnqueueDueJobs() http.HandlerFunc
	RetryJob() http.HandlerFunc
	GetOne() http.HandlerFunc
	GetJobTimeline() http.HandlerFunc
	RegisterPublicRoutes(router *mux.Router)
	RegisterPrivateRoutes(router *mux.Router)
}
