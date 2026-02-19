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
	var inserted bool
	sql := `INSERT INTO shortys (uuid, short_url, original_url) 
			VALUES ($1, $2, $3) 
			ON CONFLICT (original_url) 
			    DO UPDATE SET short_url = EXCLUDED.short_url 
			RETURNING (xmax = 0) AS inserted;`
	err = tx.QueryRow(ctx, sql, shortys.UUID, shortys.ShortURL, shortys.OriginalURL).Scan(&inserted)
	if err != nil {
		errTx := tx.Rollback(ctx)
		if errTx != nil {
			return errTx
		}
		return err
	}
	if !inserted {
		errTx := tx.Rollback(ctx)
		if errTx != nil {
			return errTx
		}

		return model.ErrConflict
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
	return model.Shortys{}
}
