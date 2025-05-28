// internal/handler/http/post_handler.go
package http

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"reddit-ingestion/internal/scraper"
)

type PostHandler struct {
	svc scraper.ScraperService
}

func NewPostHandler(svc scraper.ScraperService) *PostHandler {
	return &PostHandler{svc: svc}
}

// GetPostInfo godoc
// @Summary Get a Reddit post with comments
// @Description Retrieves a post and its comment tree from Reddit
// @Tags post
// @Accept json
// @Produce json
// @Param post_id query string true "Reddit post ID"
// @Success 200 {object} models.PostDetail
// @Failure 400 {object} models.HTTPError
// @Failure 502 {object} models.HTTPError
// @Router /post [get]
func (h *PostHandler) GetPostInfo(c echo.Context) error {
    pid := c.QueryParam("post_id")
    if pid == "" {
        return echo.NewHTTPError(http.StatusBadRequest, "missing `post_id` parameter")
    }

    ctx, cancel := context.WithTimeout(c.Request().Context(), 300*time.Second)
    defer cancel()

    detail, err := h.svc.ScrapePost(ctx, pid)
    if err != nil {
        return echo.NewHTTPError(http.StatusBadGateway, err.Error())
    }
    return c.JSON(http.StatusOK, detail)
}