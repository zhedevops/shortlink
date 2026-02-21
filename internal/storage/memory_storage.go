package storage

import (
	"context"

	"github.com/zhedevops/shortlink/internal/model"
)

type MemoryStorage struct {
	Store map[string]string
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
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
