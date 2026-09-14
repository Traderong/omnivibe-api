package database

import (
	"context"
	"fmt"
	"time"

	"github.com/Traderong/omnivibe-api/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	if cfg == nil {
		return nil, fmt.Errorf("database configuration is required")
	}

	connConfig, err := pgxpool.ParseConfig("postgres://localhost:5432/postgres")
	if err != nil {
		return nil, fmt.Errorf("create database configuration: %w", err)
	}

	connConfig.ConnConfig.Host = cfg.DBHost
	connConfig.ConnConfig.Port = cfg.DBPort
	connConfig.ConnConfig.User = cfg.DBUser
	connConfig.ConnConfig.Password = cfg.DBPassword
	connConfig.ConnConfig.Database = cfg.DBName

	connConfig.MaxConns = 20
	connConfig.MinConns = 2

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	DB = pool

	return pool, nil
}
