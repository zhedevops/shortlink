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

	if err := logger.Initialize(cnf.LogLevel); err != nil {
		return err
	}

	var completion func()

	var st repository.Repository
	dsn := strings.TrimSpace(cnf.DatabaseDsn)
	fsp := strings.TrimSpace(cnf.FileStoragePath)
	if dsn != "" {
		pool, err := database.ConnectDB(dsn)
		if err != nil {
			return err
		}
		st = storage.NewDBStorage(pool)
		completion = func() {
			log.Println("database pool closed")
			database.CloseDB(pool)
		}
	} else if fsp != "" {
		st = storage.NewFileStorage(fsp)
		completion = func() {}
	} else {
		st = storage.NewMemoryStorage()
		completion = func() {}
	}

	srv := service.NewService(st)
	h := handler.NewHandler(srv, cnf)

	if err := router.Serve(h); err != nil {
		return err
	}

	completion()

	return nil
}
