package configs

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDBConnection() (*pgxpool.Pool, error) {
	// Construir la cadena de conexión
	connString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_DATABASE"),
	)

	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Conectar al pool de conexiones
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
