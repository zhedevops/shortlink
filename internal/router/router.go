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

func NewRouter(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.GzipHandle)
	r.Get("/{id}", h.GetLinkByIDHandler)
	r.Get("/ping", h.PingHandler)
	r.With(middleware.Auth).Get("/api/user/urls", h.UserLinksHandler)
	r.With(middleware.RequireContentType("text/plain")).Post("/", h.CreateShortLinkHandler)
	r.With(middleware.RequireContentType("application/json")).Post("/api/shorten", h.CreateShortLinkEncHandler)
	r.With(middleware.RequireContentType("application/json")).Post("/api/shorten/batch", h.CreateShortLinkBatchHandler)
	return r
}

func Serve(h *handler.Handler) error {
	router := NewRouter(h)
	server := &http.Server{
		Addr:    h.Cfg.ServerAddr.ServerAddress,
		Handler: router,
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
