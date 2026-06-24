package app

import (
	artisthandler "lumiere/internal/artist/handlers"
	artistrepo "lumiere/internal/artist/repository"
	artistsvc "lumiere/internal/artist/service"
	"lumiere/internal/config"
	"lumiere/internal/database"
	homehandler "lumiere/internal/home/handlers"
	homesvc "lumiere/internal/home/service"
	lyricshandler "lumiere/internal/lyrics/handlers"
	lyricsrepo "lumiere/internal/lyrics/repository"
	lyricssvc "lumiere/internal/lyrics/service"
	playlisthandler "lumiere/internal/playlist/handlers"
	playlistrepo "lumiere/internal/playlist/repository"
	playlistsvc "lumiere/internal/playlist/service"
	userhandler "lumiere/internal/user/handlers"
	userrepo "lumiere/internal/user/repository"
	usersvc "lumiere/internal/user/service"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func New() (*echo.Echo, error) {
	_ = godotenv.Load()

	cfg, err := config.NewFromEnv()
	if err != nil {
		return nil, err
	}

	db, err := database.Connect(cfg)
	if err != nil {
		return nil, err
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World! ==> ")
	})

	api := e.Group("/api")

	userRepo := userrepo.NewGormRepo(db)
	userSvc := usersvc.New(userRepo, cfg.JWTSecret)
	userHandler := userhandler.New(userSvc)
	userGroup := api.Group("/user")
	userhandler.UserRoutes(userGroup, userHandler)

	artistRepo := artistrepo.NewGormRepo(db)
	artistSvc := artistsvc.New(artistRepo)
	artistHandler := artisthandler.New(artistSvc)
	artistGroup := api.Group("/artist")
	artisthandler.ArtistRoutes(artistGroup, artistHandler)

	lyricsRepo := lyricsrepo.NewGormRepo(db)
	lyricsSvc := lyricssvc.New(lyricsRepo)
	lyricsHandler := lyricshandler.New(lyricsSvc, artistSvc, userSvc)
	lyricsGroup := api.Group("/lyrics")
	lyricshandler.RegisterRoutes(lyricsGroup, lyricsHandler)

	playlistRepo := playlistrepo.NewGormRepo(db)
	playlistSvc := playlistsvc.New(playlistRepo, lyricsSvc)
	playlistHandler := playlisthandler.New(playlistSvc, userSvc)
	playlistGroup := api.Group("/playlist")
	playlisthandler.RegisterRoutes(playlistGroup, playlistHandler)

	homeSvc := homesvc.New(lyricsSvc, playlistSvc)
	homeHandler := homehandler.New(homeSvc)
	homeGroup := api.Group("/home")
	homehandler.RegisterRoutes(homeGroup, homeHandler)

	for _, r := range e.Routes() {
		e.Logger.Infof("route: %s %s", r.Method, r.Path)
	}

	return e, nil
}

func DefaultListenAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}

	if addr := os.Getenv("APP_HOST"); addr != "" {
		return addr
	}

	if port := os.Getenv("APP_PORT"); port != "" {
		return ":" + port
	}

	return ":4000"
}
