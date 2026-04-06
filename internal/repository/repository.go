// Package repository Обеспечивает создание, получение, редактирование и удаление информации по ссылкам.
package repository

import (
	"context"

	"github.com/zhedevops/shortlink/internal/model"
)

// Repository Интерфейс хранилища информации о ссылках.
type Repository interface {
	SetShortURL(ctx context.Context, shortys *model.Shorty) error
	GetOriginalURL(ctx context.Context, id string) model.Shorty
	CheckIDByURL(url string) model.Shorty
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context) (model.User, error)
	GetShortysByUser(ctx context.Context, userID uint32) ([]*model.Shorty, error)
	DeleteLinks(ctx context.Context, ids []string, userID uint32) error
}
