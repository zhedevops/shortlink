package model

import "time"

type Links struct {
	URL string
	ID  string
}

type URLMap struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Shortys struct {
	ID          int       `json:"id"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

func NewLinks(url string, id string) *Links {
	return &Links{
		URL: url,
		ID:  id,
	}
}
