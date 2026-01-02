package router

import (
	"fmt"
	"net/http"

	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/handler"
	"github.com/zhedevops/shortlink/internal/service"
	"github.com/zhedevops/shortlink/internal/storage"
)

func newRouter() *http.ServeMux {
	ms := storage.NewMemoryStorage()
	srv := service.NewService(ms)
	h := handler.NewHandler(srv)
	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", h.GetLinkByIDHandler)
	mux.HandleFunc("/", h.MainHandler)
	return mux
}

func Serve() error {
	cnf := config.GetConfig()
	router := newRouter()
	return http.ListenAndServe(fmt.Sprintf(`%s:%s`, cnf.HTTPURL, cnf.HTTPPort), router)
}
