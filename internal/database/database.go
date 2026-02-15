package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	if err = InitPostgres(); err != nil {
		return err
	}

	return nil
}

func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}

func InitPostgres() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	_, err := Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS shortys (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			short_url VARCHAR(8) NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_shortys_short_url 
			ON shortys(short_url);

		CREATE INDEX IF NOT EXISTS idx_shortys_original_url 
			ON shortys(original_url);
	`)

	return err
}
