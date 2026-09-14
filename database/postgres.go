package database

import (
	"context"
	"os"
	"strconv"
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
		portNumber, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, err
		}

		config.ConnConfig.Port = uint16(portNumber)
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
