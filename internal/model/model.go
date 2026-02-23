package model

import (
	"errors"
	"time"

	guid "github.com/google/uuid"
)

type Shorty struct {
	UUID        string    `json:"uuid"`
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

type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

var ErrConflict = errors.New("data conflict")

func NewShortys(uuid string, url string, id string) *Shorty {
	if uuid == "" {
		uuid = guid.New().String()
	}
	return &Shorty{
		UUID:        uuid,
		OriginalURL: url,
		ShortURL:    id,
	}
}
