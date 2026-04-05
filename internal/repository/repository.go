// Package repository Обеспечивает создание, получение, редактирование и удаление информации по ссылкам.
package repository

import (
	"context"

	"github.com/zhedevops/shortlink/internal/model"
)

// Repository Интерфейс хранилища информации о ссылках.
type Repository interface {
	SetShortURL(shortys *model.Shorty) error
	GetOriginalURL(id string) model.Shorty
	CheckIDByURL(url string) model.Shorty
	Ping(ctx context.Context) error
	CreateUser() (model.User, error)
	GetShortysByUser(userID uint32) ([]*model.Shorty, error)
	DeleteLinks(ids []string, userID uint32) error
}
