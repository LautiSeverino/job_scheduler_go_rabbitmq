package mq

import (
	"encoding/json"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"job_scheduler_go_rabbitmq/internal/core/ports"
	"log"
	"os"

	"github.com/streadway/amqp"
)

type RabbitClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
}

var _ ports.IRabbitMQClient = (*RabbitClient)(nil)

// Close implements ports.RabbitMQClient.
func (r *RabbitClient) Close() error {
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Consume implements ports.RabbitMQClient.
func (r *RabbitClient) Consume(handler func(domain.RabbitJobMessage)) error {
	msgs, err := r.channel.Consume(
		r.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		var job domain.RabbitJobMessage
		if err := json.Unmarshal(msg.Body, &job); err != nil {
			continue
		}
		handler(job)
	}

	return nil
}

// Publish implements ports.RabbitMQClient.
func (r *RabbitClient) Publish(msg domain.RabbitJobMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return r.channel.Publish(
		"",
		r.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func NewRabbitClient() (*RabbitClient, error) {
	rabbitURL := os.Getenv("RABBITMQ_URL")

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"jobs_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	log.Println("[RabbitMQ] Connected")

	return &RabbitClient{
		conn:    conn,
		channel: ch,
		queue:   q,
	}, nil
}
