package main

import (
	artisthandler "lumiere/internal/artist/handlers"
	artistrepo "lumiere/internal/artist/repository"
	artistsvc "lumiere/internal/artist/service"
	"lumiere/internal/config"
	"lumiere/internal/database"
	lyricshandler "lumiere/internal/lyrics/handlers"
	lyricsrepo "lumiere/internal/lyrics/repository"
	lyricssvc "lumiere/internal/lyrics/service"
	userhandler "lumiere/internal/user/handlers"
	userrepo "lumiere/internal/user/repository"
	usersvc "lumiere/internal/user/service"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Load .env for local development
	_ = godotenv.Load()

	cfg, err := config.NewFromEnv()
	if err != nil {
		e.Logger.Fatal(err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		e.Logger.Fatal(err)
	}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! ==> ")
	})

	// API grouping: /api
	api := e.Group("/api")

	// user feature
	userRepo := userrepo.NewGormRepo(db)
	userSvc := usersvc.New(userRepo, cfg.JWTSecret)
	userHandler := userhandler.New(userSvc)
	userGroup := api.Group("/user")
	userhandler.UserRoutes(userGroup, userHandler)

	// artist feature
	artistRepo := artistrepo.NewGormRepo(db)
	artistSvc := artistsvc.New(artistRepo)
	artistHandler := artisthandler.New(artistSvc)
	artistGroup := api.Group("/artist")
	artisthandler.ArtistRoutes(artistGroup, artistHandler)

	// lyrics feature
	lyricsRepo := lyricsrepo.NewGormRepo(db)
	lyricsSvc := lyricssvc.New(lyricsRepo)
	lyricsHandler := lyricshandler.New(lyricsSvc, artistSvc)
	lyricsGroup := api.Group("/lyrics")
	lyricshandler.RegisterRoutes(lyricsGroup, lyricsHandler)

	// Print registered routes for debugging
	for _, r := range e.Routes() {
		e.Logger.Infof("route: %s %s", r.Method, r.Path)
	}

	e.Logger.Fatal(e.Start(cfg.AppHost))
}
