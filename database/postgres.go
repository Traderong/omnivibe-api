package database

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect() (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig("postgres://localhost:5432/omnivibe")
	if err != nil {
		return nil, err
	}

	config.ConnConfig.User = os.Getenv("DB_USER")
	config.ConnConfig.Password = os.Getenv("DB_PASSWORD")
	config.ConnConfig.Database = os.Getenv("DB_NAME")

	if host := os.Getenv("DB_HOST"); host != "" {
		config.ConnConfig.Host = host
	}

	if port := os.Getenv("DB_PORT"); port != "" {
		config.ConnConfig.Port = 5432
	}

	config.MaxConns = 20
	config.MinConns = 2

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	DB = pool

	return pool, nil
}