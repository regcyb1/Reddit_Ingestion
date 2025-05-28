package parser_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"
	
	"reddit-ingestion/internal/parser"
)

func TestParseSubreddit(t *testing.T) {
	p := parser.NewRedditParser()
	ctx := context.Background()
	
	// Create test data directly in the test
	data := []byte(`{
		"data": {
			"children": [
				{
					"kind": "t3",
					"data": {
						"id": "abc123",
						"title": "Test post",
						"selftext": "This is a test post",
						"author": "testuser",
						"score": 42,
						"created_utc": 1620000000,
						"subreddit": "test",
						"permalink": "/r/test/comments/abc123/test_post",
						"url": "https://reddit.com/r/test/comments/abc123/test_post"
					}
				}
			],
			"after": "t3_next123"
		}
	}`)
	
	posts, after, err := p.ParseSubreddit(ctx, json.RawMessage(data))
	if err != nil {
		t.Fatalf("Failed to parse subreddit: %v", err)
	}
	
	if len(posts) == 0 {
		t.Error("Expected posts, got none")
	}
	
	if after == "" {
		t.Error("Expected pagination cursor, got empty string")
	}
	
	if posts[0].ID != "abc123" {
		t.Errorf("Expected post ID 'abc123', got '%s'", posts[0].ID)
	}
	
	if posts[0].Title != "Test post" {
		t.Errorf("Expected post title 'Test post', got '%s'", posts[0].Title)
	}
}

func TestParseUserInfo(t *testing.T) {
	p := parser.NewRedditParser()
	ctx := context.Background()
	
	// Create test data directly in the test
	data := []byte(`{
		"data": {
			"name": "testuser",
			"created_utc": 1620000000,
			"link_karma": 100,
			"comment_karma": 200
		}
	}`)
	
	userInfo, err := p.ParseUserInfo(ctx, json.RawMessage(data))
	if err != nil {
		t.Fatalf("Failed to parse user info: %v", err)
	}
	
	if userInfo.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", userInfo.Username)
	}
	
	if userInfo.LinkKarma != 100 {
		t.Errorf("Expected link karma 100, got %d", userInfo.LinkKarma)
	}
	
	if userInfo.CommentKarma != 200 {
		t.Errorf("Expected comment karma 200, got %d", userInfo.CommentKarma)
	}
	
	expectedTime := time.Unix(1620000000, 0)
	if !userInfo.CreatedAt.Equal(expectedTime) {
		t.Errorf("Expected creation time %v, got %v", expectedTime, userInfo.CreatedAt)
	}
}