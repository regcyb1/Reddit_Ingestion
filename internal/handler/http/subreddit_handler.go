// internal/handler/http/subreddit_handler.go
package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"reddit-ingestion/internal/scraper"
)

type SubredditHandler struct {
	svc scraper.ScraperService
}

func NewSubredditHandler(svc scraper.ScraperService) *SubredditHandler {
	return &SubredditHandler{svc: svc}
}
// GetSubredditPosts godoc
// @Summary Get posts from a subreddit
// @Description Retrieves posts from the specified subreddit with optional filters
// @Tags subreddit
// @Accept json
// @Produce json
// @Param subreddit query string true "Subreddit name without the r/ prefix"
// @Param since_timestamp query int false "Unix timestamp to filter posts"
// @Param limit query int false "Maximum number of posts to retrieve"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.HTTPError
// @Failure 502 {object} models.HTTPError
// @Router /subreddit [get]
func (h *SubredditHandler) GetSubredditPosts(c echo.Context) error {
	sr := c.QueryParam("subreddit")
	if sr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing `subreddit` parameter")
	}

	var sinceTimestamp int64
	if s := c.QueryParam("since_timestamp"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid `since_timestamp`")
		}
		sinceTimestamp = v
	}

	var limit int
	if l := c.QueryParam("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid `limit`")
		}
		limit = v
	}
	
	ctx, cancel := context.WithTimeout(c.Request().Context(), 60*time.Second)
	defer cancel()

	startTime := time.Now()

	posts, err := h.svc.ScrapeSubreddit(ctx, sr, sinceTimestamp, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("scrape error: %v", err))
	}

	duration := time.Since(startTime)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"posts": posts,
		"meta": map[string]interface{}{
			"requested_limit":    limit,
			"actual_count":       len(posts),
			"subreddit":          sr,
			"since_timestamp":    sinceTimestamp,
			"processing_time_ms": duration.Milliseconds(),
		},
	})
}