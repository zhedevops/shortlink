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

func (ms *MemoryStorage) GetOriginalURL(id string) (string, bool) {
	url, ok := ms.Store[id]
	return url, ok
}
