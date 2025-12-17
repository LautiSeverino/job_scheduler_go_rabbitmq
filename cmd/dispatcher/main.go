package main

import (
	"context"
	"log"
	"os"
	"time"

	"job_scheduler_go_rabbitmq/internal/configs"
	"job_scheduler_go_rabbitmq/internal/infra/driven/repositories"
	"job_scheduler_go_rabbitmq/internal/infra/driver/dispatcher"
	"job_scheduler_go_rabbitmq/internal/infra/driver/mq"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[DISPATCHER] .env not loaded")
	}

	log.Println("[DISPATCHER] Starting...")
	log.Println("[DISPATCHER] RABBITMQ_URL =", os.Getenv("RABBITMQ_URL"))

	// DB
	pool, err := configs.NewDBConnection()
	if err != nil {
		log.Fatal("DB error:", err)
	}
	uow := repositories.NewDataStore(pool)

	// Rabbit
	rabbit, err := mq.NewRabbitClient()
	if err != nil {
		log.Fatal("RabbitMQ error:", err)
	}

	// Dispatcher
	nodeID := os.Getenv("INSTANCE_ID")
	if nodeID == "" {
		nodeID = "dispatcher-1"
	}

	jobRepo := uow.Job()
	d := dispatcher.New(jobRepo, rabbit, nodeID)

	ctx := context.Background()

	for {
		if err := d.RunOnce(ctx); err != nil {
			log.Println("[DISPATCHER] Error:", err)
		}

		time.Sleep(1 * time.Second)
	}
}
