package service

import (
	"context"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"

	"github.com/google/uuid"
)

type JobService struct {
	uow ports.IUnitOfWork
}

func NewJobService(uow ports.IUnitOfWork) ports.IJobService {
	return &JobService{
		uow: uow,
	}
}

// Create implements ports.IJobService.
func (j *JobService) Create(ctx context.Context, input domain.CreateJobInput) (*domain.Job, error) {
	panic("unimplemented")
}

// EnqueueDueJobs implements ports.IJobService.
func (j *JobService) EnqueueDueJobs(ctx context.Context) error {
	panic("unimplemented")
}

// GetJobTimeline implements ports.IJobService.
func (j *JobService) GetJobTimeline(ctx context.Context, jobID uuid.UUID) ([]domain.JobEvent, error) {
	panic("unimplemented")
}

// GetOne implements ports.IJobService.
func (j *JobService) GetOne(ctx context.Context, params domain.JobSearchParams) (*domain.Job, error) {
	panic("unimplemented")
}

// LockNextJob implements ports.IJobService.
func (j *JobService) LockNextJob(ctx context.Context, workerID string) (*domain.Job, error) {
	panic("unimplemented")
}

// MarkJobFailed implements ports.IJobService.
func (j *JobService) MarkJobFailed(ctx context.Context, jobID uuid.UUID, errMsg string, httpStatus *int) error {
	panic("unimplemented")
}

// MarkJobRunning implements ports.IJobService.
func (j *JobService) MarkJobRunning(ctx context.Context, jobID uuid.UUID) error {
	panic("unimplemented")
}

// MarkJobSuccess implements ports.IJobService.
func (j *JobService) MarkJobSuccess(ctx context.Context, jobID uuid.UUID) error {
	panic("unimplemented")
}

// MoveToDeadLetter implements ports.IJobService.
func (j *JobService) MoveToDeadLetter(ctx context.Context, jobID uuid.UUID, reason string) error {
	panic("unimplemented")
}

// RetryJob implements ports.IJobService.
func (j *JobService) RetryJob(ctx context.Context, jobID uuid.UUID) error {
	panic("unimplemented")
}
