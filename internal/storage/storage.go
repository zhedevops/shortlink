package storage

type MemoryStorage struct {
	Store map[string]string
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		Store: make(map[string]string),
	}
}

func (ms *MemoryStorage) SetShortURL(id string, url string) {
	ms.Store[id] = url
}

func (ms *MemoryStorage) GetOriginalURL(id string) string {
	return ms.Store[id]
}

func (ms *MemoryStorage) CheckIDByURL(url string) string {
	for id, origURL := range ms.Store {
		if origURL == url {
			return id
		}
	}
	return ""
}
