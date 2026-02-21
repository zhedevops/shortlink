package repository

import (
	"context"

	"github.com/zhedevops/shortlink/internal/model"
)

type Repository interface {
	SetShortURL(shortys *model.Shorty) error
	GetOriginalURL(id string) model.Shorty
	CheckIDByURL(url string) model.Shorty
	Ping(ctx context.Context) error
}
