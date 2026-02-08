package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func ConnectDB(dsn string) error {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return err
	}

	if err = pool.Ping(ctx); err != nil {
		return err
	}

	Pool = pool
	return nil
}

func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}
