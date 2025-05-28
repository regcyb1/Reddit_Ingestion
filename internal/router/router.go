// internal/router/router.go
package router

import (
	"reddit-ingestion/internal/handler/http"
	"reddit-ingestion/internal/scraper"

	"github.com/labstack/echo/v4"
)

func NewRouter(e *echo.Echo, svc scraper.ScraperService) {
	sub := http.NewSubredditHandler(svc)
	usr := http.NewUserHandler(svc)
	pst := http.NewPostHandler(svc)
	sch := http.NewSearchHandler(svc)

	e.GET("/subreddit", sub.GetSubredditPosts)
	e.GET("/user", usr.GetUserInfo)
	e.GET("/post", pst.GetPostInfo)
	e.GET("/search", sch.Search)
}