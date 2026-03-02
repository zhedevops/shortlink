package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhedevops/shortlink/internal/model"
	"github.com/zhedevops/shortlink/internal/storage"
)

func TestServiceFuncs(t *testing.T) {
	fileName := "../../data/files/defaultpath/test.json"
	defer func() {
		_ = os.Remove(fileName)
	}()
	fs := storage.NewFileStorage(fileName)
	srv := NewService(fs)
	url := "https://example.com"
	var user = model.User{
		ID: 1,
	}
	var user2 = model.User{
		ID: 2,
	}
	var shortys = model.NewShortys("", url, "", user.ID)
	var value string

	t.Run("test CreateShortLink from url", func(t *testing.T) {
		link, err := srv.CreateShortLink(shortys)
		assert.Nil(t, err)
		assert.NotNil(t, link)
		assert.Equal(t, url, link.OriginalURL)
		assert.Len(t, link.ShortURL, 8)
	})

	t.Run("success test GetOriginalURL", func(t *testing.T) {
		link, err := srv.CreateShortLink(shortys)
		assert.Nil(t, err)
		gotURL, err := srv.GetOriginalURL(link.ShortURL)
		assert.Nil(t, err)
		assert.Equal(t, url, gotURL)
	})

	t.Run("success test getShort", func(t *testing.T) {
		id, err := srv.getShort("http://test.com", 0)
		assert.Nil(t, err)
		assert.Equal(t, "CZAqzwap", id)
	})

	t.Run("failure length test GetOriginalURL", func(t *testing.T) {
		gotURL, err := srv.GetOriginalURL("ZZZZ")
		assert.NotNil(t, err)
		assert.Equal(t, "unexpected length id", err.Error())
		assert.Empty(t, gotURL)
	})

	t.Run("failure test GetOriginalURL", func(t *testing.T) {
		gotURL, err := srv.GetOriginalURL("ZZZZZZZZ")
		assert.NotNil(t, err)
		assert.Equal(t, "url not found", err.Error())
		assert.Empty(t, gotURL)
	})

	t.Run("success test GetAuthCookie", func(t *testing.T) {
		value = srv.GetAuthCookie(user)
		assert.NotNil(t, value)
		assert.Contains(t, value, ".")
	})

	t.Run("success test CheckAuthCookie", func(t *testing.T) {
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    value,
			Path:     "/",
			HttpOnly: true,
		}
		mu, err := srv.CheckAuthCookie(cookie)
		assert.Nil(t, err)
		assert.Equal(t, mu.ID, user.ID)
	})

	t.Run("bad cookie value test CheckAuthCookie", func(t *testing.T) {
		vc := strings.ReplaceAll(value, ".", "")
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    vc,
			Path:     "/",
			HttpOnly: true,
		}
		mu, err := srv.CheckAuthCookie(cookie)
		assert.NotNil(t, err)
		assert.Equal(t, "bad cookie value", err.Error())
		assert.Empty(t, mu)
	})

	t.Run("decode cookie value failed test CheckAuthCookie", func(t *testing.T) {
		values := strings.Split(value, ".")
		data := values[0][:len(values[0])-3]
		newVc := data + "." + values[1]
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    newVc,
			Path:     "/",
			HttpOnly: true,
		}
		mu, err := srv.CheckAuthCookie(cookie)
		assert.NotNil(t, err)
		assert.Equal(t, "decode cookie value failed", err.Error())
		assert.Empty(t, mu)
	})

	t.Run("decode cookie value signature failed test CheckAuthCookie", func(t *testing.T) {
		values := strings.Split(value, ".")
		sign := values[1][:len(values[1])-3]
		newVc := values[0] + "." + sign
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    newVc,
			Path:     "/",
			HttpOnly: true,
		}
		mu, err := srv.CheckAuthCookie(cookie)
		assert.NotNil(t, err)
		assert.Equal(t, "decode cookie value signature failed", err.Error())
		assert.Empty(t, mu)
	})

	t.Run("signature verification failed test CheckAuthCookie", func(t *testing.T) {
		value2 := srv.GetAuthCookie(user2)
		values2 := strings.Split(value2, ".")
		values := strings.Split(value, ".")
		newVc := values2[0] + "." + values[1]
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    newVc,
			Path:     "/",
			HttpOnly: true,
		}
		mu, err := srv.CheckAuthCookie(cookie)
		assert.NotNil(t, err)
		assert.Equal(t, "signature verification failed", err.Error())
		assert.Empty(t, mu)
	})

	t.Run("unmarshal user data failed test CheckAuthCookie", func(t *testing.T) {
		userData, _ := json.Marshal([]byte(`{invalid json}`))
		h := hmac.New(sha256.New, secretkey)
		h.Write(userData)
		sign := h.Sum(nil)
		newVc := base64.StdEncoding.EncodeToString(userData) + "." + base64.StdEncoding.EncodeToString(sign)
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    newVc,
			Path:     "/",
			HttpOnly: true,
		}
		mu, err := srv.CheckAuthCookie(cookie)
		assert.NotNil(t, err)
		assert.Equal(t, "unmarshal user data failed", err.Error())
		assert.Empty(t, mu)
	})

	t.Run("user expired test CheckAuthCookie", func(t *testing.T) {
		// В структуре model.User нет ни UID, ни Exp, поэтому ждём user expired
		userData, _ := json.Marshal(user2)
		h := hmac.New(sha256.New, secretkey)
		h.Write(userData)
		sign := h.Sum(nil)
		newVc := base64.StdEncoding.EncodeToString(userData) + "." + base64.StdEncoding.EncodeToString(sign)
		cookie := &http.Cookie{
			Name:     "Authorization",
			Value:    newVc,
			Path:     "/",
			HttpOnly: true,
		}
		mu, err := srv.CheckAuthCookie(cookie)
		assert.NotNil(t, err)
		assert.Equal(t, "user expired", err.Error())
		assert.Empty(t, mu)
	})
}
