package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/middleware"
)

func NewRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.With(middleware.RequireMethod(http.MethodGet)).Get("/{id}", h.GetLinkByIDHandler)
	r.With(
		middleware.RequireMethod(http.MethodPost),
		middleware.RequireContentType("text/plain"),
	).Post("/", h.MainHandler)
	return r
}

func Serve(h *handler.Handler) error {
	cnf := config.GetConfig()
	router := NewRouter(h)
	return http.ListenAndServe(fmt.Sprintf(`%s:%s`, cnf.HTTPURL, cnf.HTTPPort), router)
}
