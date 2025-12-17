package worker

import (
	"context"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"log"
)

type Worker struct {
	service ports.IJobExecutionService
	rabbit  ports.IRabbitMQClient
}

func New(service ports.IJobExecutionService, rabbit ports.IRabbitMQClient) *Worker {
	return &Worker{
		service: service,
		rabbit:  rabbit,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	return w.rabbit.Consume(func(msg domain.RabbitJobMessage) {
		if err := w.service.ProcessJobMessage(ctx, msg); err != nil {
			log.Printf("[WORKER] failed processing job %s: %v", msg.JobID, err)
		}
	})
}
