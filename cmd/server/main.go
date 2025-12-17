package main

import (
	"context"
	"job_scheduler_go_rabbitmq/internal/configs"
	"job_scheduler_go_rabbitmq/internal/core/service"
	"job_scheduler_go_rabbitmq/internal/infra/driven/repositories"
	"job_scheduler_go_rabbitmq/internal/infra/driver/http/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load envirnonment variables
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file")
	}
	if os.Getenv("JOBS_SERVICE_PORT") == "" {
		log.Println("No se encontro variable de entorno JOBS_SERVICE_PORT")
	}

	//Conneccion DB
	pool, err := configs.NewDBConnection()
	if err != nil {
		log.Fatal("Database error:", err)
	}

	// Crear el repositorio de datos
	uow := repositories.NewDataStore(pool)

	// Crear el router
	router := mux.NewRouter()

	//Job
	jobService := service.NewJobService(uow, nil, nil)
	jobHandler := handler.NewJobHandler(jobService)
	handler.RegisterJobRoutes(router, jobHandler)

	// Config para manejar las señales del sistema (graceful shutdown)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar el servidor HTTP
	server := &http.Server{
		Addr:    "0.0.0.0:8000",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error en el servidor: %v", err)
		}
	}()

	log.Println("Server is running...")

	// Esperar una señal de interrupción o terminación
	<-stop

	log.Println("Shutting down server...")

	// Intentar un apagado suave del servidor con un timeout de 5 segundos
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Error en el apagado del servidor: %v", err)
	}

	log.Println("Server stopped gracefully")
}
