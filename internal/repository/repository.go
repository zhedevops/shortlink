package repository

type Repository interface {
	SetShortURL(id string, url string)
	GetOriginalURL(id string) string
	CheckIDByURL(id string) string
}
