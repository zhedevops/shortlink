package storage

import (
	"encoding/json"
	"os"
)

type FileStorage struct {
	filepath string
}

type File struct {
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

	files, err := getFiles(file)
	if err != nil {
		return
	}

	files = append(files, File{
		UUID:        len(files) + 1,
		ShortURL:    id,
		OriginalURL: url,
	})

	data, err := json.MarshalIndent(files, "", "  ")
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

	files, err := getFiles(file)
	if err != nil {
		return ""
	}

	for _, f := range files {
		if f.ShortURL == id {
			return f.OriginalURL
		}
	}
	return ""
}

func (fs *FileStorage) CheckIDByURL(url string) string {
	file, err := OpenFileStorage(fs)
	if err != nil {
		return ""
	}

	files, err := getFiles(file)
	if err != nil {
		return ""
	}

	for _, f := range files {
		if f.OriginalURL == url {
			return f.ShortURL
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

func getFiles(file *os.File) ([]File, error) {
	var files []File
	data, err := os.ReadFile(file.Name())
	if err == nil && len(data) > 0 {
		err = json.Unmarshal(data, &files)
	}
	return files, err
}
