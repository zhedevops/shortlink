package storage

import "github.com/zhedevops/shortlink/internal/model"

type MemoryStorage struct {
	Store map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Store: make(map[string]string),
	}
}

func (ms *MemoryStorage) SetShortURL(shortys *model.Shortys) error {
	ms.Store[shortys.ShortURL] = shortys.OriginalURL
	return nil
}

func (ms *MemoryStorage) GetOriginalURL(id string) model.Shortys {
	return model.Shortys{
		OriginalURL: ms.Store[id],
	}
}

func (ms *MemoryStorage) CheckIDByURL(url string) model.Shortys {
	for id, origURL := range ms.Store {
		if origURL == url {
			return model.Shortys{
				ShortURL: id,
			}
		}
	}
	return model.Shortys{}
}
