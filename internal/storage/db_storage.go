package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
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
	sql := `INSERT INTO shortys (uuid, short_url, original_url) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (original_url) 
			    DO NOTHING
			RETURNING uuid;`
	err = tx.QueryRow(ctx, sql, shortys.UUID, shortys.ShortURL, shortys.OriginalURL).Scan(new(string))
	if errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)
		return model.ErrConflict
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
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
	return model.Shortys{}
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	return dbs.db.Ping(ctx)
}
