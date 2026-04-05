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
	UserID      uint32    `json:"user_id"`
	DeletedFlag bool      `json:"is_deleted"`
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

type ResponseUserLinks struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type User struct {
	ID        uint32    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserJWT struct {
	UID uint32 `json:"uid"`
	Exp int64  `json:"exp"`
}

var (
	ErrConflict   = errors.New("data conflict")
	ErrURLDeleted = errors.New("url is deleted")
)

func NewShortys(uuid string, url string, id string, userID uint32) *Shorty {
	if uuid == "" {
		uuid = guid.New().String()
	}
	return &Shorty{
		UUID:        uuid,
		OriginalURL: url,
		ShortURL:    id,
		UserID:      userID,
	}
}
