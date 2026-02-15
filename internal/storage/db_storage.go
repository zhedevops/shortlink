package storage

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhedevops/shortlink/internal/model"
)

type DBStorage struct {
	db *pgxpool.Pool
}

func NewDBStorage(pool *pgxpool.Pool) *DBStorage {
	return &DBStorage{
		db: pool,
	}
}

func (dbs *DBStorage) SetShortURL(shortys *model.Shortys) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	tx, err := dbs.db.Begin(ctx)
	if err != nil {
		return err
	}
	sql := `INSERT INTO shortys (uuid, short_url, original_url) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, sql, shortys.UUID, shortys.ShortURL, shortys.OriginalURL)
	if err != nil {
		errTx := tx.Rollback(ctx)
		if errTx != nil {
			return errTx
		}
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (dbs *DBStorage) GetOriginalURL(id string) model.Shortys {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	row, err := dbs.db.Query(ctx, `SELECT * FROM shortys WHERE short_url = $1`, id)
	if err != nil {
		return model.Shortys{}
	}
	defer row.Close()
	var shortys model.Shortys
	if row.Next() {
		err = row.Scan(
			&shortys.UUID,
			&shortys.ShortURL,
			&shortys.OriginalURL,
			&shortys.CreatedAt)
		if err != nil {
			return model.Shortys{}
		}
	}
	return shortys
}

func (dbs *DBStorage) CheckIDByURL(url string) model.Shortys {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	row, err := dbs.db.Query(ctx, `SELECT * FROM shortys WHERE original_url = $1`, url)
	if err != nil {
		return model.Shortys{}
	}
	defer row.Close()
	var shortys model.Shortys
	if row.Next() {
		err = row.Scan(
			&shortys.UUID,
			&shortys.ShortURL,
			&shortys.OriginalURL,
			&shortys.CreatedAt,
		)
		if err != nil {
			return model.Shortys{}
		}
	}
	return shortys
}
