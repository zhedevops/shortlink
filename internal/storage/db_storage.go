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

func (dbs *DBStorage) SetShortURL(shortys *model.Shorty) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	tx, err := dbs.db.Begin(ctx)
	if err != nil {
		return err
	}
	sql := `INSERT INTO shortys (uuid, short_url, original_url, user_id) 
			VALUES ($1, $2, $3, $4) 
			ON CONFLICT (original_url) 
			    DO NOTHING
			RETURNING uuid;`
	err = tx.QueryRow(ctx, sql, shortys.UUID, shortys.ShortURL, shortys.OriginalURL, shortys.UserID).Scan(new(string))
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

func (dbs *DBStorage) GetOriginalURL(id string) model.Shorty {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	row, err := dbs.db.Query(ctx, `SELECT * FROM shortys WHERE short_url = $1`, id)
	if err != nil {
		return model.Shorty{}
	}
	defer row.Close()
	var shortys model.Shorty
	if row.Next() {
		err = row.Scan(
			&shortys.UUID,
			&shortys.ShortURL,
			&shortys.OriginalURL,
			&shortys.CreatedAt,
			&shortys.UserID,
			&shortys.DeletedFlag)
		if err != nil {
			return model.Shorty{}
		}
	}
	return shortys
}

func (dbs *DBStorage) CheckIDByURL(url string) model.Shorty {
	return model.Shorty{}
}

func (dbs *DBStorage) Ping(ctx context.Context) error {
	return dbs.db.Ping(ctx)
}

func (dbs *DBStorage) CreateUser() (model.User, error) {
	var user = model.User{}
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	tx, err := dbs.db.Begin(ctx)
	if err != nil {
		return user, err
	}
	err = tx.QueryRow(ctx, `INSERT INTO users DEFAULT VALUES RETURNING id, created_at;`).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		_ = tx.Rollback(ctx)
		return user, err
	}
	return user, tx.Commit(ctx)
}

func (dbs *DBStorage) GetShortysByUser(userID uint32) ([]*model.Shorty, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
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

func (dbs *DBStorage) DeleteLinks(ids []string, userID uint32) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	tx, err := dbs.db.Begin(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE shortys SET is_deleted = true WHERE short_url = ANY($1) AND user_id = $2`, ids, userID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
