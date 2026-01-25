package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/middleware"
)

func NewRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/{id}", h.GetLinkByIDHandler)
	r.With(middleware.RequireContentType("text/plain")).Post("/", h.CreateShortLinkHandler)
	r.With(middleware.RequireContentType("application/json")).Post("/api/shorten", h.CreateShortLinkEncHandler)
	return r
}

func Serve(h *handler.Handler) error {
	router := NewRouter(h)
	return http.ListenAndServe(h.Cfg.ServerAddr.ServerAddress, router)
}
