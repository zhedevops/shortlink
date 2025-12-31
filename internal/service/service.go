package service

import (
	"crypto/sha1"

	"github.com/zhedevops/shortlink/internal/model"
)

// 52 буквы
const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var store = make(map[string]string)

func CreateShortLink(urlStr string) *model.Links {
	id := getShort(urlStr)
	store[id] = urlStr
	return model.NewLinks(urlStr, id)
}

func getShort(url string) string {
	hash := sha1.Sum([]byte(url))
	b := hash[:8]
	res := make([]byte, 8)
	for i := 0; i < 8; i++ {
		res[i] = chars[b[i]%52]
	}
	return string(res)
}

func GetOriginalURL(id string) (string, bool) {
	url, ok := store[id]
	return url, ok
}
