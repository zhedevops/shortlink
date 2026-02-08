package main

import (
	"log"

	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/database"
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
	config.SetConfig()
	cnf := config.GetConfig()
	if err := database.ConnectDB(cnf.DatabaseDsn); err != nil {
		return err
	}
	defer database.CloseDB()
	if err := logger.Initialize(cnf.LogLevel); err != nil {
		return err
	}
	fs, err := storage.NewFileStorage(cnf.FileStoragePath)
	if err != nil {
		return err
	}
	srv := service.NewService(fs)
	h := handler.NewHandler(srv, cnf)
	return router.Serve(h)
}
