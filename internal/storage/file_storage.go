package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/zhedevops/shortlink/internal/model"
)

type FileStorage struct {
	filepath string
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func NewFileStorage(filepath string) *FileStorage {
	return &FileStorage{
		filepath: filepath,
	}
}

func (fs *FileStorage) SetShortURL(shortys *model.Shortys) error {
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

	data, err := json.Marshal(shortys)
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

func (fs *FileStorage) GetOriginalURL(id string) model.Shortys {
	return findMatchingElement(fs.filepath, true, id)
}

func (fs *FileStorage) CheckIDByURL(url string) model.Shortys {
	return findMatchingElement(fs.filepath, false, url)
}

func findMatchingElement(filepath string, isShorten bool, searchValue string) model.Shortys {
	file, err := os.Open(filepath)
	if err != nil {
		return model.Shortys{}
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

		var shortys model.Shortys
		if err := json.Unmarshal([]byte(line), &shortys); err != nil {
			continue
		}
		if isShorten {
			if shortys.ShortURL == searchValue {
				return shortys
			}
		} else {
			if shortys.OriginalURL == searchValue {
				return shortys
			}
		}
	}

	return model.Shortys{}
}
