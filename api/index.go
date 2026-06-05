package main

import (
	"net/http"
	"sync"

	"lumiere/internal/app"

	"github.com/labstack/echo/v4"
)

var (
	serverOnce sync.Once
	server     *echo.Echo
	serverErr  error
)

func Handler(w http.ResponseWriter, r *http.Request) {
	serverOnce.Do(func() {
		server, serverErr = app.New()
	})

	if serverErr != nil {
		http.Error(w, serverErr.Error(), http.StatusInternalServerError)
		return
	}

	server.ServeHTTP(w, r)
}
