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
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, err
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
