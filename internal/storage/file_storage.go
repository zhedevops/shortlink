package storage

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/zhedevops/shortlink/internal/model"
)

type FileStorage struct {
	filepath string
}

func NewFileStorage(filepath string) *FileStorage {
	return &FileStorage{
		filepath: filepath,
	}
}

func (fs *FileStorage) SetShortURL(id string, url string) error {
	file, err := os.OpenFile(fs.filepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	if _, err = file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	var next int
	if count == 0 {
		next++
	} else {
		next = count - 1
	}

	um := model.URLMap{
		UUID:        next,
		ShortURL:    id,
		OriginalURL: url,
	}

	data, err := json.Marshal(um)
	if err != nil {
		return err
	}

	if next == 1 {
		_, err = file.WriteString("[\n  ")
		if err != nil {
			return err
		}
		_, err = file.Write(data)
		if err != nil {
			return err
		}
		_, err = file.WriteString("\n]")
		if err != nil {
			return err
		}

		return nil
	}

	fi, _ := file.Stat()
	_ = file.Truncate(fi.Size() - 2)
	_, _ = file.Seek(0, io.SeekEnd)
	_, _ = file.Write([]byte(",\n  "))
	_, _ = file.Write(data)
	_, _ = file.Write([]byte("\n]"))

	return nil
}

func (fs *FileStorage) GetOriginalURL(id string) string {
	return findMatchingElement(fs.filepath, true, id)
}

func (fs *FileStorage) CheckIDByURL(url string) string {
	return findMatchingElement(fs.filepath, false, url)
}

func findMatchingElement(filepath string, isShorten bool, searchValue string) string {
	file, err := os.Open(filepath)
	if err != nil {
		return ""
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "[" || line == "]" || line == "" {
			continue
		}
		line = strings.TrimSuffix(line, ",")

		var um model.URLMap
		if err := json.Unmarshal([]byte(line), &um); err != nil {
			continue
		}
		if isShorten {
			if um.ShortURL == searchValue {
				return um.OriginalURL
			}
		} else {
			if um.OriginalURL == searchValue {
				return um.ShortURL
			}
		}
	}

	return ""
}
