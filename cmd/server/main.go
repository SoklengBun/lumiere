package main

import (
	"log"
	"net/http"
	"os"

	"lumiere/internal/app"
)

func main() {
	e, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("PORT")
	if addr != "" {
		addr = ":" + addr
	} else {
		addr = app.DefaultListenAddr()
	}

	e.Logger.Infof("listening on %s", addr)
	if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
		e.Logger.Fatal(err)
	}
}
