// Package storage Хранилище данных
package storage

import (
	"context"
	"errors"

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

func (dbs *DBStorage) SetShortURL(ctx context.Context, shortys *model.Shorty) error {
	sql := `INSERT INTO shortys (uuid, short_url, original_url, user_id) 
			VALUES ($1, $2, $3, $4) 
			ON CONFLICT (original_url) 
			    DO NOTHING
			RETURNING uuid;`
	var uuid string
	err := dbs.db.QueryRow(ctx, sql, shortys.UUID, shortys.ShortURL, shortys.OriginalURL, shortys.UserID).Scan(&uuid)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.ErrConflict
	}
	return err
}

func (dbs *DBStorage) GetOriginalURL(ctx context.Context, id string) *model.Shorty {
	row := dbs.db.QueryRow(ctx, `SELECT uuid, short_url, original_url, created_at, user_id, is_deleted FROM shortys WHERE short_url = $1`, id)
	var shortys model.Shorty
	if err := row.Scan(
		&shortys.UUID,
		&shortys.ShortURL,
		&shortys.OriginalURL,
		&shortys.CreatedAt,
		&shortys.UserID,
		&shortys.DeletedFlag); err != nil {
		return &model.Shorty{}
	}
	return &shortys
}

func (dbs *DBStorage) CheckIDByURL(url string) *model.Shorty {
	_ = url
	return &model.Shorty{}
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	return dbs.db.Ping(ctx)
}

func (dbs *DBStorage) CreateUser(ctx context.Context) (model.User, error) {
	user := model.User{}
	tx, err := dbs.db.Begin(ctx)
	if err != nil {
		return user, err
	}
	if err := tx.QueryRow(
		ctx,
		`INSERT INTO users DEFAULT VALUES RETURNING id, created_at;`,
	).Scan(&user.ID, &user.CreatedAt); err != nil {
		_ = tx.Rollback(ctx)
		return user, err
	}
	return user, tx.Commit(ctx)
}

func (dbs *DBStorage) GetShortysByUser(ctx context.Context, userID uint32) ([]*model.Shorty, error) {
	var shortys []*model.Shorty
	rows, err := dbs.db.Query(ctx, `SELECT short_url, original_url FROM shortys WHERE user_id = $1`, userID)
	if err != nil {
		return shortys, err
	}
	defer rows.Close()
	for rows.Next() {
		var shorty model.Shorty
		err = rows.Scan(&shorty.ShortURL, &shorty.OriginalURL)
		if err != nil {
			return shortys, err
		}
		shortys = append(shortys, &shorty)
	}
	return shortys, nil
}

func (dbs *DBStorage) DeleteLinks(ctx context.Context, ids []string, userID uint32) error {
	tx, err := dbs.db.Begin(ctx)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(
		ctx,
		`UPDATE shortys SET is_deleted = true WHERE short_url = ANY($1) AND user_id = $2`,
		ids,
		userID,
	); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (dbs *DBStorage) GetStats(ctx context.Context) (model.ResponseStats, error) {
	stats := model.ResponseStats{}
	row := dbs.db.QueryRow(ctx, `SELECT count(short_url) as urls, count(DISTINCT user_id) as users FROM shortys`)
	if err := row.Scan(&stats.URLs, &stats.Users); err != nil {
		return stats, err
	}

	return stats, nil
}
