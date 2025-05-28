package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/labstack/echo/v4"
	handler "reddit-ingestion/internal/handler/http"
	"reddit-ingestion/internal/models"
)

type MockScraperService struct {
	ScrapeSubredditFunc   func(ctx context.Context, subreddit string, sinceTimestamp int64, limit int) ([]models.Post, error)
	ScrapeUserActivityFunc func(ctx context.Context, username string, sinceTimestamp int64, postLimit, commentLimit int) (models.UserActivity, error)
	ScrapePostFunc        func(ctx context.Context, postID string) (models.PostDetail, error)
	SearchFunc            func(ctx context.Context, searchParams map[string]string, sinceTimestamp int64, limit int) ([]models.Post, error)
}

func (m *MockScraperService) ScrapeSubreddit(ctx context.Context, subreddit string, sinceTimestamp int64, limit int) ([]models.Post, error) {
	return m.ScrapeSubredditFunc(ctx, subreddit, sinceTimestamp, limit)
}

func (m *MockScraperService) ScrapeUserActivity(ctx context.Context, username string, sinceTimestamp int64, postLimit, commentLimit int) (models.UserActivity, error) {
	return m.ScrapeUserActivityFunc(ctx, username, sinceTimestamp, postLimit, commentLimit)
}

func (m *MockScraperService) ScrapePost(ctx context.Context, postID string) (models.PostDetail, error) {
	return m.ScrapePostFunc(ctx, postID)
}

func (m *MockScraperService) Search(ctx context.Context, searchParams map[string]string, sinceTimestamp int64, limit int) ([]models.Post, error) {
	return m.SearchFunc(ctx, searchParams, sinceTimestamp, limit)
}

func TestSubredditHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/subreddit?subreddit=test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	
	mockService := &MockScraperService{
		ScrapeSubredditFunc: func(ctx context.Context, subreddit string, sinceTimestamp int64, limit int) ([]models.Post, error) {
			return []models.Post{
				{
					ID:     "123",
					Title:  "Test Post",
					Author: "testuser",
				},
			}, nil
		},
	}
	
	h := handler.NewSubredditHandler(mockService)
	if err := h.GetSubredditPosts(c); err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	posts, ok := response["posts"].([]interface{})
	if !ok || len(posts) != 1 {
		t.Errorf("Expected 1 post in response, got %v", posts)
	}
}
