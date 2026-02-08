package database

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func ConnectDb(dsn string) error {
	if strings.TrimSpace(dsn) == "" {
		return errors.New("dsn is empty")
	}

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

func CloseDb() {
	if Pool != nil {
		Pool.Close()
	}
}
