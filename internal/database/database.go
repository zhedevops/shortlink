package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
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

	if err := InitPostgres(dsn); err != nil {
		return err
	}

	return nil
}

func CloseDB() {
	if Pool != nil {
		Pool.Close()
	}
}

func InitPostgres(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() {
		_ = db.Close()
	}()

	if err = goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err = goose.Up(db, "../../migrations"); err != nil {
		return err
	}

	return nil
}
