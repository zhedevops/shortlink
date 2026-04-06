// Package model Содержит описание структуры модели ссылки, метода создания модели и основных типов.
package model

import (
	"errors"
	"time"

	guid "github.com/google/uuid"
)

// Shorty Базовая модель ссылки. Состоит из:
//   - идентификатора
//   - короткой ссылки
//   - оригинальной ссылки
//   - даты создания
//   - идентификатора пользователя, создавшего ссылку
//   - флага пометки в качестве удалённой записи.
type Shorty struct {
	// UUID Идентификатор
	UUID string `json:"uuid"`
	// ShortURL Короткая ссылка
	ShortURL string `json:"short_url"`
	// OriginalURL Оригинальная ссылка
	OriginalURL string `json:"original_url"`
	// CreatedAt Дата создания
	CreatedAt time.Time `json:"created_at"`
	// UserID Идентификатор создателя
	UserID uint32 `json:"user_id"`
	// DeletedFlag Флаг пометки ссылки в качестве удалённой
	DeletedFlag bool `json:"is_deleted"`
}

// Request Запрос, содержащий в теле json с оригинальной ссылкой.
type Request struct {
	URL string `json:"url"`
}

// Response json-ответ с результатом конвертации оригинальной ссылки.
type Response struct {
	Result string `json:"result"`
}

// RequestBatch Запрос, содержащий в теле json с множеством оригинальных ссылок и идентификатором для формирования ответа и привязки короткой ссылки к оригинальной.
type RequestBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResponseBatch json-ответ с результатом пакетной конвертации оригинальных ссылок.
type ResponseBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ResponseUserLinks json-ответ, содержащий короткую и оригинальную ссылки.
type ResponseUserLinks struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// ErrorResponse Сообщение об ошибке конфертации.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// User Тип пользователя, содержащий идентификатор и дату создания.
type User struct {
	ID        uint32    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

// UserJWT Тип, описывающий JWT пользователя, содержащий уникальный идентификатор и дату окончания действия JWT.
type UserJWT struct {
	UID uint32 `json:"uid"`
	Exp int64  `json:"exp"`
}

var (
	ErrConflict               = errors.New("data conflict")
	ErrURLDeleted             = errors.New("url is deleted")
	ErrEmptyURL               = errors.New("empty url")
	ErrWrongID                = errors.New("unexpected length id")
	ErrURLNotFound            = errors.New("url not found")
	ErrBadCookie              = errors.New("bad cookie value")
	ErrDecodeCookie           = errors.New("decode cookie value failed")
	ErrDecodeCookieSignature  = errors.New("decode cookie value signature failed")
	ErrSignatureVerification  = errors.New("signature verification failed")
	ErrUnmarshal              = errors.New("unmarshal user data failed")
	ErrExpired                = errors.New("user expired")
	ErrGenerateURL            = errors.New("failed to generate unique short url with max attempts")
	ErrServerAddressFlagValue = errors.New("need url in a form protocol:host:port")
	ErrHostPort               = errors.New("host or port is empty")
)

// NewShortys Создаёт новую базовую модель ссылки.
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
