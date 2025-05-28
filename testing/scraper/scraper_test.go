package scraper_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"
	
	"reddit-ingestion/internal/models"
	"reddit-ingestion/internal/scraper"
	"reddit-ingestion/testing/mocks"
)

func TestScrapeSubreddit(t *testing.T) {
	mockClient := &mocks.MockRedditClient{}
	mockParser := &mocks.MockParser{}
	
	// Setup test data
	mockPosts := []models.Post{
		{
			ID:        "abcd123",
			Title:     "Test Post",
			Body:      "Test Body",
			Author:    "testuser",
			Score:     100,
			CreatedAt: time.Now(),
			URL:       "https://reddit.com/r/test/comments/abcd123",
		},
	}
	
	// Define mock behaviors with more specific control
	mockClient.GetSubredditURLFunc = func(subreddit string, limit int, after string) string {
		return "https://reddit.com/r/" + subreddit + "/new.json"
	}
	
	// Track how many times FetchJSON is called
	fetchCount := 0
	mockClient.FetchJSONFunc = func(ctx context.Context, url string) (json.RawMessage, error) {
		fetchCount++
		return json.RawMessage(`{"data":{"children":[]}}`), nil
	}
	
	// Make sure ParseSubreddit only returns our posts once, then empty results
	mockParser.ParseSubredditFunc = func(ctx context.Context, data json.RawMessage) ([]models.Post, string, error) {
		if fetchCount == 1 {
			return mockPosts, "", nil // No "after" pagination token to prevent looping
		}
		return []models.Post{}, "", nil
	}
	
	// Create service with mocks
	svc := scraper.NewScraperService(mockClient, mockParser)
	
	// Test the service - explicitly set limit to 1 to control behavior
	posts, err := svc.ScrapeSubreddit(context.Background(), "test", 0, 1)
	if err != nil {
		t.Fatalf("Failed to scrape subreddit: %v", err)
	}
	
	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}
	
	if len(posts) > 0 && posts[0].ID != "abcd123" {
		t.Errorf("Expected post ID 'abcd123', got '%s'", posts[0].ID)
	}
}