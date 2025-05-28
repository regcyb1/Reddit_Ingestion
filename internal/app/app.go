// internal/app/app.go
package app

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"reddit-ingestion/internal/client"
	"reddit-ingestion/internal/config"
	"reddit-ingestion/internal/parser"
	"reddit-ingestion/internal/router"
	"reddit-ingestion/internal/scraper"
)

type App struct {
	Config  *config.Config
	Echo    *echo.Echo
	Service scraper.ScraperService
	Client  *client.RedditClient
	Parser  parser.Parser
}

func Initialize() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	redditClient, err := client.NewRedditClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Reddit client: %w", err)
	}
	
	redditParser := parser.NewRedditParser()
	scraperService := scraper.NewScraperService(redditClient, redditParser)
	
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	
	router.NewRouter(e, scraperService)
	
	return &App{
		Config:  cfg,
		Echo:    e,
		Service: scraperService,
		Client:  redditClient,
		Parser:  redditParser,
	}, nil
}

func (a *App) Start() error {
	port := a.Config.ServerPort
	if port == "" {
		port = "8080"
	}
	return a.Echo.Start(":" + port)
}