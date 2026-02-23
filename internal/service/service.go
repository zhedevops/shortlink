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

func NewService(r repository.Repository) *Service {
	return &Service{repo: r}
}

func (srv *Service) CreateShortLink(shortys *model.Shorty) (*model.Shorty, error) {
	urlStr := strings.TrimSpace(shortys.OriginalURL)
	if len(urlStr) == 0 {
		return nil, errors.New("empty url")
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
	id, err := srv.getShort(urlStr, 0)
	if err != nil {
		return nil, fmt.Errorf("failed create short link: %w", err)
	}
	shortys.ShortURL = id
	err = srv.repo.SetShortURL(shortys)
	if err != nil {
		if errors.Is(err, model.ErrConflict) {
			return shortys, err
		}
		return nil, fmt.Errorf("failed set short link: %w", err)
	}
	return shortys, nil
}

func (srv *Service) getShort(url string, attempt int) (string, error) {
	if attempt >= maxAttempts {
		return "", errors.New("failed to generate unique short url with max attempts")
	}
	hash := sha1.Sum([]byte(url))
	b := hash[:8]
	res := make([]byte, 8)
	for i := 0; i < 8; i++ {
		res[i] = chars[b[i]%52]
	}
	strID := string(res)
	existingShortys := srv.repo.GetOriginalURL(strID)
	if existingShortys.OriginalURL != "" && existingShortys.OriginalURL != url {
		ns := strconv.FormatInt(time.Now().UnixNano(), 10)
		return srv.getShort(url+ns, attempt+1)
	}
	return strID, nil
}

func (srv *Service) GetOriginalURL(id string) (string, error) {
	if len(id) != 8 {
		return "", errors.New("unexpected length id")
	}
	existingShortys := srv.repo.GetOriginalURL(id)
	if existingShortys.OriginalURL == "" {
		return "", errors.New("url not found")
	}
	return existingShortys.OriginalURL, nil
}

func (srv *Service) Ping(ctx context.Context) error {
	return srv.repo.Ping(ctx)
}

func (srv *Service) GetNewUser() (model.User, error) {
	return srv.repo.CreateUser()
}

func (srv *Service) CheckAuthCookie(cookieAuth *http.Cookie) (model.User, error) {
	user := model.User{}
	ujwt := model.UserJWT{}
	values := strings.Split(cookieAuth.Value, ".")
	if len(values) != 2 {
		return user, errors.New("bad cookie value")
	}
	jwtData, err := base64.StdEncoding.DecodeString(values[0])
	if err != nil {
		return user, errors.New("decode cookie value failed")
	}
	signature, err := base64.StdEncoding.DecodeString(values[1])
	if err != nil {
		return user, errors.New("decode cookie value signature failed")
	}
	h := hmac.New(sha256.New, secretkey)
	h.Write(jwtData)
	sign := h.Sum(nil)
	if !hmac.Equal(sign, signature) {
		return user, errors.New("signature verification failed")
	}
	err = json.Unmarshal(jwtData, &ujwt)
	if err != nil {
		return user, errors.New("unmarshal user data failed")
	}
	if ujwt.Exp < time.Now().Unix() {
		return user, errors.New("user expired")
	}
	user.ID = ujwt.UID
	return user, nil
}

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

func (srv *Service) GetUserLinks(userID uint32) ([]*model.Shorty, error) {
	return srv.repo.GetShortysByUser(userID)
}
