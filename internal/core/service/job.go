package service

import (
	"context"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"sort"

	"github.com/google/uuid"
)

type JobService struct {
	uow    ports.IUnitOfWork
	exec   ports.IJobExecutor
	rabbit ports.IRabbitMQClient
}

func NewJobService(uow ports.IUnitOfWork, exec ports.IJobExecutor, rabbit ports.IRabbitMQClient) *JobService {
	return &JobService{
		uow:    uow,
		exec:   exec,
		rabbit: rabbit,
	}
}

var _ ports.IJobService = (*JobService)(nil)
var _ ports.IJobExecutionService = (*JobService)(nil)

func (s *JobService) ProcessJobMessage(
	ctx context.Context,
	msg domain.RabbitJobMessage,
) error {

	var retryMsg *domain.RabbitJobMessage

	err := s.uow.Atomic(ctx, func(uow ports.IUnitOfWork) error {

		//  Load job
		job, err := uow.Job().GetOne(ctx, domain.JobSearchParams{
			ID: &msg.JobID,
		})
		if err != nil {
			return err
		}

		// Idempotencia
		if job.Status != domain.JobStatusQueued {
			return nil
		}

		//  Mark running
		if err := uow.Job().MarkRunning(ctx, job.ID); err != nil {
			return err
		}

		//  Ejecutar callback (lado técnico)
		result := s.exec.Execute(ctx, job)
		// SUCCESS
		if result.Error == nil {
			attempt := domain.NewAttempt(
				job.ID,
				msg.Attempt,
				domain.AttemptStatusSuccess,
				nil,
				&result.HTTPStatus,
			)

			if err := uow.Attempt().Insert(ctx, attempt); err != nil {
				return err
			}

			if err := uow.Job().MarkCompleted(ctx, job.ID); err != nil {
				return err
			}

			if err := uow.Event().Insert(
				ctx,
				domain.NewJobSucceededEvent(job.ID),
			); err != nil {
				return err
			}

			return nil
		}

		//  FAILED
		errMsg := result.Error.Error()

		attempt := domain.NewAttempt(
			job.ID,
			msg.Attempt,
			domain.AttemptStatusFailed,
			&errMsg,
			&result.HTTPStatus,
		)

		if err := uow.Attempt().Insert(ctx, attempt); err != nil {
			return err
		}

		if msg.Attempt < job.MaxRetries {
			if err := uow.Job().MarkFailed(ctx, job.ID, errMsg, nil); err != nil {
				return err
			}

			if err := uow.Event().Insert(
				ctx,
				domain.NewJobFailedEvent(job.ID, errMsg),
			); err != nil {
				return err
			}

			next := msg
			next.Attempt++
			retryMsg = &next
			return nil
		}

		//  DEAD
		if err := uow.Job().MarkDead(ctx, job.ID, "max retries exceeded"); err != nil {
			return err
		}

		return uow.Event().Insert(
			ctx,
			domain.NewJobDeadEvent(job.ID),
		)
	})

	if err != nil {
		return err
	}

	if retryMsg != nil {
		return s.rabbit.Publish(*retryMsg)
	}

	return nil
}

// Create implements ports.IJobService.
func (s *JobService) Create(ctx context.Context, input domain.CreateJobInput) (*domain.Job, error) {
	job, err := domain.NewJob(input)
	if err != nil {
		return nil, err
	}

	err = s.uow.Atomic(ctx, func(d ports.IUnitOfWork) error {
		err = d.Job().Insert(ctx, *job)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return job, nil
}

// GetOne implements ports.IJobService.
func (s *JobService) GetOne(ctx context.Context, params domain.JobSearchParams) (*domain.Job, error) {
	job, err := s.uow.Job().GetOne(ctx, params)
	if err != nil {
		return nil, err
	}
	return job, nil
}

// GetTimeline implements ports.IJobService.
func (s *JobService) GetTimeline(ctx context.Context, jobID uuid.UUID) ([]domain.Event, error) {
	events, err := s.uow.Event().Get(ctx, domain.EventSearchParams{
		JobID: &jobID,
	})
	if err != nil {
		return nil, err
	}

	attempts, err := s.uow.Attempt().Get(ctx, domain.AttemptSearchParams{
		JobID: &jobID,
	})
	if err != nil {
		return nil, err
	}

	for _, a := range attempts {
		var t domain.EventType
		var msg string

		if a.Status == domain.AttemptStatusSuccess {
			t = domain.EventJobSucceeded
			msg = "attempt succeeded"
		} else {
			t = domain.EventJobFailed
			msg = "attempt failed"
		}

		events = append(events, domain.Event{
			ID:        uuid.New(),
			JobID:     jobID,
			Type:      t,
			Message:   msg,
			CreatedAt: a.CreatedAt,
		})
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})

	return events, nil
}
