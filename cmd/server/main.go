package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load envirnonment variables
	err := godotenv.Load()
	if err != nil {
		log.Print("Error loading .env file")
	}
	if os.Getenv("ESPORTS_SERVICE_PORT") == "" {
		log.Println("No se encontro variable de entorno ESPORTS_SERVICE_PORT")
	}
}
