package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zhedevops/shortlink/internal/storage"
)

func TestServiceFuncs(t *testing.T) {
	ms := storage.NewMemoryStorage()
	srv := NewService(ms)
	url := "https://example.com"

	t.Run("test CreateShortLink from url", func(t *testing.T) {
		link := srv.CreateShortLink(url)
		assert.NotNil(t, link)
		assert.Equal(t, url, link.URL)
		assert.Len(t, link.ID, 8)
	})

	t.Run("success test GetOriginalURL", func(t *testing.T) {
		link := srv.CreateShortLink(url)
		gotURL, ok := srv.GetOriginalURL(link.ID)
		assert.True(t, ok)
		assert.Equal(t, url, gotURL)
	})

	t.Run("failure test GetOriginalURL", func(t *testing.T) {
		gotURL, ok := srv.GetOriginalURL("ZZZZZZZZ")
		assert.False(t, ok)
		assert.Empty(t, gotURL)
	})
}
