package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/service"
)

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "expected POST method", http.StatusBadRequest)
		return
	}
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/plain") {
		http.Error(w, "unsupported content type", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	urlStr := strings.TrimSpace(string(body))
	if len(urlStr) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}
	u, err := url.ParseRequestURI(urlStr)
	if err != nil {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		http.Error(w, "unsupported scheme", http.StatusBadRequest)
		return
	}

	link := service.CreateShortLink(urlStr)
	cnf := config.GetConfig()
	resp := fmt.Sprintf("http://%s:%s/%s\r\n", cnf.HTTPURL, cnf.HTTPPort, link.ID)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(resp))
	if err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func GetLinkByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "expected GET method", http.StatusBadRequest)
		return
	}
	id := r.URL.Path[1:]
	urlStr, ok := service.GetOriginalURL(id)
	if !ok {
		http.Error(w, "url not found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", urlStr)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
