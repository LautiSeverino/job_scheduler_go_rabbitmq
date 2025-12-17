package ports

import "job_scheduler_go_rabbitmq/internal/core/domain"

type IRabbitMQClient interface {
	Publish(msg domain.RabbitJobMessage) error
	Consume(handler func(domain.RabbitJobMessage)) error
	Close() error
}
