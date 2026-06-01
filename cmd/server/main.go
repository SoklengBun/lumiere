package main

import (
	"log"
	"net/http"

	"lumiere/internal/app"
)

func main() {
	e, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	addr := app.DefaultListenAddr()
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
