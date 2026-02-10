package main

import (
	"log"
	"strings"

	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/database"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/logger"
	"github.com/zhedevops/shortlink/internal/repository"
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
	config.SetConfig()
	cnf := config.GetConfig()

	dsn := strings.TrimSpace(cnf.DatabaseDsn)
	if dsn != "" {
		if err := database.ConnectDB(dsn); err != nil {
			return err
		}
		defer database.CloseDB()
	}

	if err := logger.Initialize(cnf.LogLevel); err != nil {
		return err
	}

	var st repository.Repository
	fsp := strings.TrimSpace(cnf.FileStoragePath)
	if fsp != "" {
		st = storage.NewFileStorage(fsp)
	} else {
		st = storage.NewMemoryStorage()
	}

	srv := service.NewService(st)
	h := handler.NewHandler(srv, cnf)

	return router.Serve(h)
}
