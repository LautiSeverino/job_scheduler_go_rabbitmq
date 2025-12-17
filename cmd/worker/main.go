package main

import (
	"context"
	"log"
	"os"

	"job_scheduler_go_rabbitmq/internal/configs"
	"job_scheduler_go_rabbitmq/internal/core/service"
	"job_scheduler_go_rabbitmq/internal/infra/driven/executor"
	"job_scheduler_go_rabbitmq/internal/infra/driven/repositories"
	"job_scheduler_go_rabbitmq/internal/infra/driver/mq"
	"job_scheduler_go_rabbitmq/internal/infra/driver/worker"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("[WORKER] Starting...")
	if err := godotenv.Load(); err != nil {
		log.Println("[WORKER] No .env file loaded")
	}

	// DB
	pool, err := configs.NewDBConnection()
	if err != nil {
		log.Fatal("DB error:", err)
	}
	uow := repositories.NewDataStore(pool)

	log.Println("[WORKER] RABBITMQ_URL =", os.Getenv("RABBITMQ_URL"))
	log.Println("[WORKER] RABBITMQ_QUEUE =", os.Getenv("RABBITMQ_QUEUE"))

	// Rabbit
	rabbit, err := mq.NewRabbitClient()
	if err != nil {
		log.Fatal("RabbitMQ error:", err)
	}

	// Executor
	exec := executor.NewHTTPExecutor()

	// Job Service (concreto)
	jobService := service.NewJobService(uow, exec, rabbit)

	// Worker
	w := worker.New(jobService, rabbit)

	ctx := context.Background()

	log.Println("[WORKER] Listening for jobs...")
	if err := w.Start(ctx); err != nil {
		log.Fatal("[WORKER] stopped:", err)
	}
}
