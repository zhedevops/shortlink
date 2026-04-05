package main

import (
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/zhedevops/shortlink/internal/audit"
	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/container"
	"github.com/zhedevops/shortlink/internal/database"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/logger"
	"github.com/zhedevops/shortlink/internal/repository"
	"github.com/zhedevops/shortlink/internal/router"
	"github.com/zhedevops/shortlink/internal/service"
	"github.com/zhedevops/shortlink/internal/storage"
)

func main() {
	// Для профилирования CPU используем этот код
	f, err := os.Create("result.pprof")
	if err != nil {
		panic(err)
	}

	_ = pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	if err := run(); err != nil {
		log.Fatal(err)
	}
	// Для профилирования потребления памяти используем этот код
	//ff, _ := os.Create("heap.pprof")
	//_ = pprof.WriteHeapProfile(ff)
	//_ = ff.Close()
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

	var sinks []audit.AuditSink
	af := strings.TrimSpace(cnf.AuditFile)
	au := strings.TrimSpace(cnf.AuditURL)
	if af != "" {
		sinks = append(sinks, &audit.FileSink{Path: af, Mu: sync.Mutex{}})
	}
	if au != "" {
		client := &http.Client{
			Timeout: 10 * time.Second,
		}
		sinks = append(sinks, &audit.RemoteSink{URL: au, Client: client})
	}
	auditSrv := audit.NewAuditService(sinks)

	app := &container.App{
		Audit:   auditSrv,
		Service: srv,
		Config:  cnf,
	}
	h := handler.NewHandler(app)

	if err := router.Serve(h); err != nil {
		return err
	}

	completion()

	return nil
}
