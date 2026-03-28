package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/zhedevops/shortlink/internal/audit"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/container"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/service"
	"github.com/zhedevops/shortlink/internal/storage"
)

func TestNewRouter(t *testing.T) {
	cnf := config.GetConfig()
	fileName := "../../data/files/defaultpath/test.json"
	defer func() {
		_ = os.Remove(fileName)
	}()
	fs := storage.NewFileStorage(fileName)
	srv := service.NewService(fs)
	var sinks []audit.AuditSink
	auditSrv := audit.NewAuditService(sinks)
	app := &container.App{
		Audit:   auditSrv,
		Service: srv,
		Config:  cnf,
	}
	h := handler.NewHandler(app)
	r := NewRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("https://ria.ru/"))
	req.Header.Add("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}
