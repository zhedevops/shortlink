// Package service Сервис, отвечающий за обработку запросов обработчика.
package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/repository"
)

// 52 буквы
const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const maxAttempts = 5

var secretkey = []byte("supersecretkey")

type Service struct {
	repo repository.Repository
}

// NewService Создаёт сервис.
func NewService(r repository.Repository) *Service {
	return &Service{repo: r}
}

// CreateShortLink Создаёт короткую ссылку.
func (srv *Service) CreateShortLink(ctx context.Context, shortys *model.Shorty) (*model.Shorty, error) {
	urlStr := strings.TrimSpace(shortys.OriginalURL)
	if len(urlStr) == 0 {
		return nil, model.ErrEmptyURL
	}
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %s", u.Scheme)
	}
	existingShortys := srv.repo.CheckIDByURL(urlStr)
	if existingShortys.ShortURL != "" {
		return &existingShortys, nil
	}
	id, err := srv.getShort(ctx, urlStr, 0)
	if err != nil {
		return nil, fmt.Errorf("failed create short link: %w", err)
	}
	shortys.ShortURL = id
	if err = srv.repo.SetShortURL(ctx, shortys); err != nil {
		if errors.Is(err, model.ErrConflict) {
			return shortys, err
		}
		return nil, fmt.Errorf("failed set short link: %w", err)
	}
	return shortys, nil
}

// GetOriginalURL Получает оригинальную ссылку.
func (srv *Service) GetOriginalURL(ctx context.Context, id string) (string, error) {
	if len(id) != 8 {
		return "", model.ErrWrongID
	}
	existingShortys := srv.repo.GetOriginalURL(ctx, id)
	if existingShortys.OriginalURL == "" {
		return "", model.ErrURLNotFound
	}
	if existingShortys.DeletedFlag {
		return "", model.ErrURLDeleted
	}
	return existingShortys.OriginalURL, nil
}

func (srv *Service) Ping(ctx context.Context) error {
	return srv.repo.Ping(ctx)
}

// GetNewUser Создаёт нового пользователя.
func (srv *Service) GetNewUser(ctx context.Context) (model.User, error) {
	return srv.repo.CreateUser(ctx)
}

// CheckAuthCookie Проверяет авторизационную cookie.
func (srv *Service) CheckAuthCookie(cookieAuth *http.Cookie) (model.User, error) {
	user := model.User{}
	ujwt := model.UserJWT{}
	values := strings.Split(cookieAuth.Value, ".")
	if len(values) != 2 {
		return user, model.ErrBadCookie
	}
	jwtData, err := base64.StdEncoding.DecodeString(values[0])
	if err != nil {
		return user, model.ErrDecodeCookie
	}
	signature, err := base64.StdEncoding.DecodeString(values[1])
	if err != nil {
		return user, model.ErrDecodeCookieSignature
	}
	h := hmac.New(sha256.New, secretkey)
	h.Write(jwtData)
	sign := h.Sum(nil)
	if !hmac.Equal(sign, signature) {
		return user, model.ErrSignatureVerification
	}
	if err = json.Unmarshal(jwtData, &ujwt); err != nil {
		return user, model.ErrUnmarshal
	}
	if ujwt.Exp < time.Now().Unix() {
		return user, model.ErrExpired
	}
	user.ID = ujwt.UID
	return user, nil
}

// GetAuthCookie Создаёт авторизационную cookie.
func (srv *Service) GetAuthCookie(user model.User) string {
	userJWT := model.UserJWT{
		UID: user.ID,
		Exp: time.Now().Add(time.Hour).Unix(),
	}
	userData, _ := json.Marshal(userJWT)
	h := hmac.New(sha256.New, secretkey)
	h.Write(userData)
	sign := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(userData) + "." + base64.StdEncoding.EncodeToString(sign)
}

// GetUserLinks Получает все ссылки пользователя.
func (srv *Service) GetUserLinks(ctx context.Context, userID uint32) ([]*model.Shorty, error) {
	return srv.repo.GetShortysByUser(ctx, userID)
}

// DeleteLinks Удаляет ссылки пользователя по списку.
func (srv *Service) DeleteLinks(ctx context.Context, URLs []string, userID uint32) error {
	inputCh := deleteLinksFanIn(URLs)
	var ids []string
	for in := range inputCh {
		ids = append(ids, in)
	}

	return srv.repo.DeleteLinks(ctx, ids, userID)
}

func (srv *Service) getShort(ctx context.Context, url string, attempt int) (string, error) {
	if attempt >= maxAttempts {
		return "", model.ErrGenerateURL
	}
	hash := sha1.Sum([]byte(url))
	b := hash[:8]
	res := make([]byte, 8)
	for i := 0; i < 8; i++ {
		res[i] = chars[b[i]%52]
	}
	strID := string(res)
	existingShortys := srv.repo.GetOriginalURL(ctx, strID)
	if existingShortys.OriginalURL != "" && existingShortys.OriginalURL != url {
		ns := strconv.FormatInt(time.Now().UnixNano(), 10)
		return srv.getShort(ctx, url+ns, attempt+1)
	}
	return strID, nil
}

func deleteLinksFanIn(URLs []string) chan string {
	inputCh := make(chan string, len(URLs))
	go func() {
		defer close(inputCh)
		for _, u := range URLs {
			inputCh <- u
		}
	}()
	return inputCh
}
