package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

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
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %s", u.Scheme)
	}
	existingID := srv.repo.CheckIDByURL(urlStr)
	if existingID != "" {
		return model.NewLinks(urlStr, existingID), nil
	}
	id := srv.getShort(urlStr)
	srv.repo.SetShortURL(id, urlStr)
	return model.NewLinks(urlStr, id), nil
}

func (srv *Service) getShort(url string) string {
	hash := sha1.Sum([]byte(url))
	b := hash[:8]
	res := make([]byte, 8)
	for i := 0; i < 8; i++ {
		res[i] = chars[b[i]%52]
	}
	strID := string(res)
	existingURL := srv.repo.GetOriginalURL(strID)
	if existingURL != "" && existingURL != url {
		ns := fmt.Sprintf("%d", time.Now().UnixNano())
		return srv.getShort(url + ns)
	}
	return strID
}

func (srv *Service) GetOriginalURL(id string) (string, error) {
	if len(id) != 8 {
		return "", errors.New("unexpected length id")
	}
	origURL := srv.repo.GetOriginalURL(id)
	if origURL == "" {
		return "", errors.New("url not found")
	}
	return origURL, nil
}
