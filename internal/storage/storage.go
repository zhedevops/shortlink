package storage

import (
	"encoding/json"
	"os"
)

type FileStorage struct {
	filepath string
}

type URLMap struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(filepath string) (*FileStorage, error) {
	return &FileStorage{
		filepath: filepath,
	}, nil
}

func (fs *FileStorage) SetShortURL(id string, url string) {
	file, err := OpenFileStorage(fs)
	if err != nil {
		return
	}

	usms, err := getURLsMaps(file)
	if err != nil {
		return
	}

	usms = append(usms, URLMap{
		UUID:        len(usms) + 1,
		ShortURL:    id,
		OriginalURL: url,
	})

	data, err := json.MarshalIndent(usms, "", "  ")
	if err != nil {
		return
	}

	err = os.WriteFile(file.Name(), data, 0666)
	if err != nil {
		return
	}
}

func (fs *FileStorage) GetOriginalURL(id string) string {
	file, err := OpenFileStorage(fs)
	if err != nil {
		return ""
	}

	usms, err := getURLsMaps(file)
	if err != nil {
		return ""
	}

	for _, um := range usms {
		if um.ShortURL == id {
			return um.OriginalURL
		}
	}
	return ""
}

func (fs *FileStorage) CheckIDByURL(url string) string {
	file, err := OpenFileStorage(fs)
	if err != nil {
		return ""
	}

	usms, err := getURLsMaps(file)
	if err != nil {
		return ""
	}

	for _, um := range usms {
		if um.OriginalURL == url {
			return um.ShortURL
		}
	}

	return ""
}

func OpenFileStorage(fs *FileStorage) (*os.File, error) {
	file, err := os.OpenFile(fs.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	return file, nil
}

func getURLsMaps(file *os.File) ([]URLMap, error) {
	var usms []URLMap
	data, err := os.ReadFile(file.Name())
	if err == nil && len(data) > 0 {
		err = json.Unmarshal(data, &usms)
	}
	return usms, err
}
