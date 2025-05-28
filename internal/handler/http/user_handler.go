// internal/handler/http/user_handler.go
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

type UserHandler struct {
	svc scraper.ScraperService
}

func NewUserHandler(svc scraper.ScraperService) *UserHandler {
	return &UserHandler{svc: svc}
}
// GetUserInfo godoc
// @Summary Get information about a Reddit user
// @Description Retrieves profile information, posts, and comments for a specific Reddit user
// @Tags user
// @Accept json
// @Produce json
// @Param username query string true "Reddit username"
// @Param since_timestamp query int false "Unix timestamp to filter posts and comments (newer than this timestamp)"
// @Param post_limit query int false "Maximum number of posts to retrieve. Use -1 for all available posts"
// @Param comment_limit query int false "Maximum number of comments to retrieve. Use -1 for all available comments"
// @Success 200 {object} models.UserActivity "Returns user information, posts, and comments"
// @Failure 400 {object} models.HTTPError "Invalid request parameters"
// @Failure 502 {object} models.HTTPError "Error occurred while scraping data"
// @Router /user [get]
func (h *UserHandler) GetUserInfo(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing `username` parameter")
	}

	var sinceTimestamp int64
	if s := c.QueryParam("since_timestamp"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid `since_timestamp`")
		}
		sinceTimestamp = v
	}

	var postLimit int
	if l := c.QueryParam("post_limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid `post_limit`")
		}
		postLimit = v
	}

	var commentLimit int
	if l := c.QueryParam("comment_limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid `comment_limit`")
		}
		commentLimit = v
	} else {
		commentLimit = postLimit
	}
	if postLimit < -1 || commentLimit < -1 {
		return echo.NewHTTPError(http.StatusBadRequest, "limits must be -1 or a positive integer")
	}

	// Increase timeout for unlimited fetching
	timeout := 60 * time.Second
	if (postLimit == -1 || commentLimit == -1) && sinceTimestamp > 0 {
		timeout = 240 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
	defer cancel()

	activity, err := h.svc.ScrapeUserActivity(ctx, username, sinceTimestamp, postLimit, commentLimit)
	if err != nil {
		return echo.NewHTTPError(
			http.StatusBadGateway,
			fmt.Sprintf("scrape user data error: %v", err),
		)
	}

	return c.JSON(http.StatusOK, activity)
}