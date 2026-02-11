package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBStorage struct {
	db *pgxpool.Pool
}

func NewDBStorage(pool *pgxpool.Pool) *DBStorage {
	return &DBStorage{
		db: pool,
	}
}

func (dbs *DBStorage) SetShortURL(id string, url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	tx, err := dbs.db.Begin(ctx)
	if err != nil {
		return err
	}
	ct, err := tx.Exec(ctx, `INSERT INTO shortys (short_url, original_url) VALUES ($1, $2)`, id, url)
	if err != nil {
		errTx := tx.Rollback(ctx)
		if errTx != nil {
			return err
		}
		return err
	}
	if ct.RowsAffected() == 0 {
		errTx := tx.Rollback(ctx)
		if errTx != nil {
			return err
		}
		return errors.New("SetShortURL: no rows inserted")
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (dbs *DBStorage) GetOriginalURL(id string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	row, err := dbs.db.Query(ctx, `SELECT original_url FROM shortys WHERE short_url = $1`, id)
	if err != nil {
		return ""
	}
	defer row.Close()
	var originalURL string
	if row.Next() {
		err = row.Scan(&originalURL)
		if err != nil {
			return ""
		}
	}
	return originalURL
}

func (dbs *DBStorage) CheckIDByURL(url string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	row, err := dbs.db.Query(ctx, `SELECT short_url FROM shortys WHERE original_url = $1`, url)
	if err != nil {
		return ""
	}
	defer row.Close()
	var shortURL string
	if row.Next() {
		err = row.Scan(&shortURL)
		if err != nil {
			return ""
		}
	}
	return shortURL
}
