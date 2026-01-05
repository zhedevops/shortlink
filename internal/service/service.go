package service

import (
	"crypto/sha1"
	"errors"
	"net/url"
	"strings"

	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/repository"
)

// 52 буквы
const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type Service struct {
	repo repository.Repository
}

func NewService(r repository.Repository) *Service {
	return &Service{repo: r}
}

func (srv *Service) CreateShortLink(urlStr string) (*model.Links, error) {
	urlStr = strings.TrimSpace(urlStr)
	if len(urlStr) == 0 {
		return nil, errors.New("empty url")
	}
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("unsupported scheme")
	}
	id := getShort(urlStr)
	srv.repo.SetShortURL(id, urlStr)
	return model.NewLinks(urlStr, id), nil
}

func getShort(url string) string {
	hash := sha1.Sum([]byte(url))
	b := hash[:8]
	res := make([]byte, 8)
	for i := 0; i < 8; i++ {
		res[i] = chars[b[i]%52]
	}
	return string(res)
}

func (srv *Service) GetOriginalURL(id string) (string, error) {
	if len(id) != 8 {
		return "", errors.New("unexpected length id")
	}
	origURL, ok := srv.repo.GetOriginalURL(id)
	if !ok {
		return "", errors.New("url not found")
	}
	return origURL, nil
}
