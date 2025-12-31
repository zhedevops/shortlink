package router

import (
	"fmt"
	"net/http"

	"github.com/zhedevops/shortlink/internal/config"
	"github.com/zhedevops/shortlink/internal/handler"
)

func newRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/{id}", handler.GetLinkByIDHandler)
	mux.HandleFunc("/", handler.MainHandler)
	return mux
}

func Serve() error {
	cnf := config.GetConfig()
	router := newRouter()
	return http.ListenAndServe(fmt.Sprintf(`%s:%s`, cnf.HTTPURL, cnf.HTTPPort), router)
}
