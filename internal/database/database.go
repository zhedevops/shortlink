// Package database Пакет БД
package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// ConnectDB Осуществляет соединение с БД
func ConnectDB(dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DSN: %w", err)
	}

	config.MaxConns = 20                        // 20-50 типично
	config.MinConns = 2                         // 2-5
	config.MaxConnLifetime = 5 * time.Minute    // 5m-1h
	config.MaxConnIdleTime = 10 * time.Minute   // 10m
	config.HealthCheckPeriod = 30 * time.Second // 30s

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	stats := pool.Stat()
	if stats.MaxConns() == 0 {
		pool.Close()
		return nil, fmt.Errorf("invalid pool config: max_conns=0")
	}

	if err = InitPostgres(pool); err != nil {
		return nil, err
	}

	return pool, nil
}

// CloseDB Закрывает соединение с БД
func CloseDB(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

// InitPostgres Запускает миграции
func InitPostgres(pool *pgxpool.Pool) error {
	env, _ := os.LookupEnv("ENVIRONMENT")
	if env == "dev" {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.UpContext(ctx, db, "./migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
