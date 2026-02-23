package service

import (
	"os"
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
	var shortys = model.NewShortys("", url, "")

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
}
