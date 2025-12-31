package main

import (
	"log"

	"github.com/zhedevops/shortlink/internal/router"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return router.Serve()
}
