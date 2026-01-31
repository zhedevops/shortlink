package storage

import (
	"encoding/json"
	"os"
)

type FileStorage struct {
	filepath string
}

type File struct {
	Uuid         int    `json:"uuid"`
	Short_URL    string `json:"short_url"`
	Original_URL string `json:"original_url"`
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
		Uuid:         len(files) + 1,
		Short_URL:    id,
		Original_URL: url,
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
		if f.Short_URL == id {
			return f.Original_URL
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
		if f.Original_URL == url {
			return f.Short_URL
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
