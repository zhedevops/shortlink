// Package router Маршрутизатор
package router

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/middleware"
)

// NewRouter Создаёт новый маршрутизатор
func NewRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.GzipHandle)
	r.Get("/{id}", h.GetLinkByIDHandler)
	r.Get("/ping", h.PingHandler)
	r.Get("/api/user/urls", h.UserLinksHandler)
	r.With(middleware.RequireContentType("text/plain")).Post("/", h.CreateShortLinkHandler)
	r.With(middleware.RequireContentType("application/json")).Post("/api/shorten", h.CreateShortLinkEncHandler)
	r.With(middleware.RequireContentType("application/json")).Post("/api/shorten/batch", h.CreateShortLinkBatchHandler)
	r.With(middleware.RequireContentType("application/json")).Delete("/api/user/urls", h.DeleteLinkBatchHandler)
	return r
}

// Serve Запускает сервис и осуществляет его корректную остановку
func Serve(h *handler.Handler) error {
	router := NewRouter(h)
	server := &http.Server{
		Addr:              h.Cfg.ServerAddr.ServerAddress,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Канал для получения сигналов прерывания
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Println("HTTP server started")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	// Когда будет получен сигнал прерывания выполнится код
	<-signalChan

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("HTTP server stoped")

	return server.Shutdown(shutdownCtx)
}
