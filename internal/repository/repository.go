package repository

import "github.com/zhedevops/shortlink/internal/model"

type Repository interface {
	SetShortURL(shortys *model.Shortys) error
	GetOriginalURL(id string) model.Shortys
	CheckIDByURL(url string) model.Shortys
}
