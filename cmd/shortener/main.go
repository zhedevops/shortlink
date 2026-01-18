package main

import (
	"log"

	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/logger"
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
	cnf := config.GetConfig()
	if err := logger.Initialize(cnf.LogLevel); err != nil {
		return err
	}
	ms := storage.NewMemoryStorage()
	srv := service.NewService(ms)
	h := handler.NewHandler(srv, cnf)
	return router.Serve(h)
}
