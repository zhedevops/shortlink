package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

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

func CloseDB(pool *pgxpool.Pool) {
	if pool != nil {
		pool.Close()
	}
}

func InitPostgres(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS shortys (
			uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			short_url VARCHAR(8) NOT NULL,
			original_url TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			user_id INT NOT NULL,
			is_deleted BOOLEAN NOT NULL DEFAULT FALSE
		);

		CREATE INDEX IF NOT EXISTS idx_shortys_short_url 
			ON shortys(short_url);

		CREATE UNIQUE INDEX IF NOT EXISTS idx_shortys_original_url 
			ON shortys(original_url);
        
        CREATE TABLE IF NOT EXISTS users (
            id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
        );
	`)

	return err
}
