package dispatcher

import (
	"context"
	"log"

	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"job_scheduler_go_rabbitmq/utils"
)

// Dispatcher is responsible for dispatching jobs to RabbitMQ.
type Dispatcher struct {
	repo   ports.IJobRepository
	rabbit ports.IRabbitMQClient
	nodeID string
}

// New creates a new Dispatcher instance.
func New(repo ports.IJobRepository, rabbit ports.IRabbitMQClient, nodeID string) *Dispatcher {
	return &Dispatcher{
		repo:   repo,
		rabbit: rabbit,
		nodeID: nodeID,
	}
}

// RunOnce dispatches pending ready jobs by publishing them to RabbitMQ.
func (d *Dispatcher) RunOnce(ctx context.Context) error {
	limit := uint(50)
	pending := domain.JobStatusPending

	jobs, err := d.repo.Get(ctx, domain.JobSearchParams{
		Status:     &pending,
		ReadyToRun: func(b bool) *bool { return &b }(true),
		LockFree:   func(b bool) *bool { return &b }(true),
		SearchParams: utils.SearchParams{
			Limit: &limit,
		},
	})
	if err != nil {
		return err
	}

	for _, job := range jobs {
		if err := d.repo.LockJob(ctx, job.ID, d.nodeID); err != nil {
			continue
		}

		msg := domain.NewRabbitJobMessageFromJob(job)

		if err := d.rabbit.Publish(msg); err != nil {
			log.Printf("[DISPATCHER] failed to publish job %s: %v", job.ID, err)
			continue
		}

		if err := d.repo.MarkQueued(ctx, job.ID); err != nil {
			log.Printf("[DISPATCHER] failed to mark job %s queued: %v", job.ID, err)
			continue
		}

		log.Printf("[DISPATCHER] Job %s queued", job.ID)
	}

	return nil
}
