package main

import (
	"log"

	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/router"
	"github.com/zhedevops/shortlink/internal/service"
	"github.com/zhedevops/shortlink/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ms := storage.NewMemoryStorage()
	srv := service.NewService(ms)
	h := handler.NewHandler(srv)
	return router.Serve(h)
}
