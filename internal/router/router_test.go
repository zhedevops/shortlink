package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/service"
	"github.com/zhedevops/shortlink/internal/storage"
)

func TestNewRouter(t *testing.T) {
	cnf := config.GetConfig()
	ms := storage.NewMemoryStorage()
	srv := service.NewService(ms)
	h := handler.NewHandler(srv, cnf)
	r := NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ria.ru/"))
	req.Header.Add("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
