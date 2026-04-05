package storage

import (
	"context"

	"github.com/zhedevops/shortlink/internal/model"
)

type MemoryStorage struct {
	Store map[string]string
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	_ = ctx
	return nil
}

func (ms *MemoryStorage) CreateUser() (model.User, error) {
	return model.User{}, nil
}

func (ms *MemoryStorage) GetShortysByUser(userID uint32) ([]*model.Shorty, error) {
	_ = userID
	return nil, nil
}

func (ms *MemoryStorage) DeleteLinks(ids []string, userID uint32) error {
	_ = ids
	_ = userID
	return nil
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Store: make(map[string]string),
	}
}

func (ms *MemoryStorage) SetShortURL(shortys *model.Shorty) error {
	ms.Store[shortys.ShortURL] = shortys.OriginalURL
	return nil
}

func (ms *MemoryStorage) GetOriginalURL(id string) model.Shorty {
	return model.Shorty{
		OriginalURL: ms.Store[id],
	}
}

func (ms *MemoryStorage) CheckIDByURL(url string) model.Shorty {
	for id, origURL := range ms.Store {
		if origURL == url {
			return model.Shorty{
				ShortURL: id,
			}
		}
	}
	return model.Shorty{}
}
