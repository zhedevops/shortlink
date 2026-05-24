// Package storage Хранилище данных в памяти
package storage

import (
	"context"
	"sync"

	"github.com/zhedevops/shortlink/internal/model"
)

type MemoryStorage struct {
	mu    sync.Mutex
	Store map[string]string
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}

func (ms *MemoryStorage) CreateUser(ctx context.Context) (model.User, error) {
	_ = ctx
	return model.User{}, nil
}

func (ms *MemoryStorage) GetShortysByUser(ctx context.Context, userID uint32) ([]*model.Shorty, error) {
	_ = ctx
	_ = userID
	return nil, nil
}

func (ms *MemoryStorage) DeleteLinks(ctx context.Context, ids []string, userID uint32) error {
	_ = ctx
	_ = ids
	_ = userID
	return nil
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		mu:    sync.Mutex{},
		Store: make(map[string]string),
	}
}

func (ms *MemoryStorage) SetShortURL(ctx context.Context, shortys *model.Shorty) error {
	_ = ctx
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.Store[shortys.ShortURL] = shortys.OriginalURL
	return nil
}

func (ms *MemoryStorage) GetOriginalURL(ctx context.Context, id string) *model.Shorty {
	_ = ctx
	return &model.Shorty{
		OriginalURL: ms.Store[id],
	}
}

func (ms *MemoryStorage) CheckIDByURL(url string) *model.Shorty {
	for id, origURL := range ms.Store {
		if origURL == url {
			return &model.Shorty{
				ShortURL: id,
			}
		}
	}
	return &model.Shorty{}
}

func (ms *MemoryStorage) GetStats(ctx context.Context) (model.ResponseStats, error) {
	_ = ctx
	return model.ResponseStats{}, nil
}
