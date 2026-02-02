package repository

type Repository interface {
	SetShortURL(id string, url string) error
	GetOriginalURL(id string) string
	CheckIDByURL(url string) string
}
