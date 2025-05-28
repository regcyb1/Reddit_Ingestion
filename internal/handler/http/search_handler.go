// internal/handler/http/search_handler.go
package http

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"reddit-ingestion/internal/scraper"

	"github.com/labstack/echo/v4"
)

type SearchHandler struct {
	svc scraper.ScraperService
}

func NewSearchHandler(svc scraper.ScraperService) *SearchHandler {
	return &SearchHandler{svc: svc}
}

// Search godoc
// @Summary Search Reddit for posts
// @Description Search Reddit with various filters and parameters
// @Tags search
// @Accept json
// @Produce json
// @Param search_string query string false "Search query string"
// @Param since_timestamp query int false "Unix timestamp to filter posts"
// @Param limit query int false "Maximum number of results"
// @Param sort query string false "Sort order (relevance, hot, top, new, comments)"
// @Param time query string false "Time range (hour, day, week, month, year, all)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.HTTPError
// @Failure 502 {object} models.HTTPError
// @Router /search [get]
func (h *SearchHandler) Search(c echo.Context) error {
	query := c.QueryParam("search_string")

	var limit int = 25 // Default
	if l := c.QueryParam("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid 'limit' parameter")
		}
		limit = v
	}

	var sinceTimestamp int64
	if s := c.QueryParam("since_timestamp"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid 'since_timestamp' parameter")
		}
		sinceTimestamp = v
	}

	// Validate parameter combinations
	if limit < -1 {
		return echo.NewHTTPError(http.StatusBadRequest, "limit must be -1 or a positive integer")
	}

	// Increase timeout for unlimited fetching
	timeout := 60 * time.Second
	if limit == -1 && sinceTimestamp > 0 {
		timeout = 240 * time.Second
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
	defer cancel()

	startTime := time.Now()

	searchParams := buildSearchParams(c)

	posts, err := h.svc.Search(ctx, searchParams, sinceTimestamp, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, fmt.Sprintf("search_string error: %v", err))
	}

	duration := time.Since(startTime)

	// Add additional metadata for unlimited fetching
	limitDescription := fmt.Sprintf("%d", limit)
	if limit == -1 {
		if sinceTimestamp > 0 {
			limitDescription = "all items since timestamp"
		} else {
			limitDescription = "default maximum (500)"
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"posts": posts,
		"meta": map[string]interface{}{
			"query":              query,
			"params":             searchParams,
			"count":              len(posts),
			"processing_time_ms": duration.Milliseconds(),
			"requested_limit":    limitDescription,
		},
	})
}

func buildSearchParams(c echo.Context) map[string]string {
	params := make(map[string]string)

	if searchString := c.QueryParam("search_string"); searchString != "" {
		params["search_string"] = searchString
	}

	// Pagination
	if after := c.QueryParam("after"); after != "" {
		params["after"] = after
	}

	if before := c.QueryParam("before"); before != "" {
		params["before"] = before
	}

	// Sorting
	if sort := c.QueryParam("sort"); sort != "" {
		params["sort"] = sort
	} else {
		params["sort"] = "relevance"
	}

	if timeRange := c.QueryParam("time"); timeRange != "" {
		params["time"] = timeRange
	} else {
		params["time"] = "all"
	}

	// Handle limit
	if limit := c.QueryParam("limit"); limit != "" {
		params["limit"] = limit
	} else {
		params["limit"] = "25"
	}

	// Advanced parameters
	advancedParams := []string{
		"subreddit", "author", "site", "url", "selftext", "self", "nsfw", "restrict_sr",
	}

	for _, param := range advancedParams {
		if value := c.QueryParam(param); value != "" {
			params[param] = value
		}
	}

	// Handle compound query if provided
	if compoundQuery := c.QueryParam("compound_query"); compoundQuery != "" {
		parts := strings.Fields(compoundQuery)
		for _, part := range parts {
			if strings.Contains(part, ":") {
				kv := strings.SplitN(part, ":", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					value := strings.TrimSpace(kv[1])

					switch key {
					case "subreddit", "author", "site", "url", "selftext", "self", "nsfw":
						params[key] = value
					}
				}
			} else {
				if params["search_string"] == "" {
					params["search_string"] = part
				} else {
					params["search_string"] = params["search_string"] + " " + part
				}
			}
		}
	}

	return params
}
