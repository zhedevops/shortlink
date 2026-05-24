// Package storage Файловое хранилище данных
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
	_ = ctx
	return nil
}

func (fs *FileStorage) CreateUser(ctx context.Context) (model.User, error) {
	_ = ctx
	return model.User{}, nil
}

func (fs *FileStorage) GetShortysByUser(ctx context.Context, userID uint32) ([]*model.Shorty, error) {
	_ = ctx
	_ = userID
	return nil, nil
}

func NewFileStorage(filepath string) *FileStorage {
	return &FileStorage{
		filepath: filepath,
	}
}

func (fs *FileStorage) DeleteLinks(ctx context.Context, ids []string, userID uint32) error {
	_ = ctx
	_ = ids
	_ = userID
	return nil
}

func (fs *FileStorage) SetShortURL(ctx context.Context, shortys *model.Shorty) error {
	_ = ctx
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
		if _, err := file.WriteString("[\n  "); err != nil {
			return err
		}
		if _, err := file.Write(data); err != nil {
			return err
		}
		if _, err := file.WriteString("\n]"); err != nil {
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

func (fs *FileStorage) GetOriginalURL(ctx context.Context, id string) *model.Shorty {
	_ = ctx
	return findMatchingElement(fs.filepath, true, id)
}

func (fs *FileStorage) CheckIDByURL(url string) *model.Shorty {
	return findMatchingElement(fs.filepath, false, url)
}

func findMatchingElement(filepath string, isShorten bool, searchValue string) *model.Shorty {
	file, err := os.Open(filepath)
	if err != nil {
		return &model.Shorty{}
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

		var shortys model.Shorty
		if err := json.Unmarshal([]byte(line), &shortys); err != nil {
			continue
		}
		if isShorten {
			if shortys.ShortURL == searchValue {
				return &shortys
			}
		} else {
			if shortys.OriginalURL == searchValue {
				return &shortys
			}
		}
	}

	return &model.Shorty{}
}

func (fs *FileStorage) GetStats(ctx context.Context) (model.ResponseStats, error) {
	_ = ctx
	return model.ResponseStats{}, nil
}
