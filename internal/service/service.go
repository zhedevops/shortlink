package service

import (
	"crypto/sha1"

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

func (srv *Service) CreateShortLink(urlStr string) *model.Links {
	id := getShort(urlStr)
	srv.repo.SetShortURL(id, urlStr)
	return model.NewLinks(urlStr, id)
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

func (srv *Service) GetOriginalURL(id string) (string, bool) {
	return srv.repo.GetOriginalURL(id)
}
