// testing/integration/integration_test.go
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"reddit-ingestion/internal/config"
	"reddit-ingestion/internal/models"
	"reddit-ingestion/internal/parser"
	"reddit-ingestion/internal/router"
	"reddit-ingestion/internal/scraper"
)

// MockableRedditClient extends the RedditClient for testing
type MockableRedditClient struct {
	MockResponse map[string]json.RawMessage
}

// Implement all methods from the RedditClientInterface

func (m *MockableRedditClient) FetchJSON(ctx context.Context, url string) (json.RawMessage, error) {
	log.Printf("MockClient: FetchJSON called with URL: %s", url)
	
	// Try exact match first
	if response, exists := m.MockResponse[url]; exists {
		log.Printf("MockClient: Found exact match for URL: %s", url)
		return response, nil
	}
	
	// Try partial match
	for mockedURL, response := range m.MockResponse {
		if strings.Contains(url, mockedURL) {
			log.Printf("MockClient: Found partial match: %s for URL: %s", mockedURL, url)
			return response, nil
		}
	}
	
	log.Printf("MockClient: No match found for URL: %s, returning default response", url)
	// Default mock response for any URL not explicitly defined
	return json.RawMessage(`{"data":{"children":[]}}`), nil
}

func (m *MockableRedditClient) FetchMoreComments(ctx context.Context, postID string, commentIDs []string) (json.RawMessage, error) {
	log.Printf("MockClient: FetchMoreComments called for post: %s", postID)
	return json.RawMessage(`{"json":{"data":{"things":[]}}}`), nil
}

func (m *MockableRedditClient) GetSubredditURL(subreddit string, limit int, after string) string {
	url := fmt.Sprintf("https://reddit.com/r/%s/new.json?raw_json=1", subreddit)
	if limit > 0 {
		url += fmt.Sprintf("&limit=%d", limit)
	}
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}
	log.Printf("MockClient: GetSubredditURL generated: %s", url)
	return url
}

func (m *MockableRedditClient) GetUserAboutURL(username string) string {
	url := fmt.Sprintf("https://reddit.com/user/%s/about.json", username)
	log.Printf("MockClient: GetUserAboutURL generated: %s", url)
	return url
}

func (m *MockableRedditClient) GetUserPostsURL(username string, after string) string {
	url := fmt.Sprintf("https://reddit.com/user/%s/submitted/new.json?raw_json=1&sort=new", username)
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}
	log.Printf("MockClient: GetUserPostsURL generated: %s", url)
	return url
}

func (m *MockableRedditClient) GetUserCommentsURL(username string, after string) string {
	url := fmt.Sprintf("https://reddit.com/user/%s/comments/.json?raw_json=1&limit=100", username)
	if after != "" {
		url += fmt.Sprintf("&after=%s", after)
	}
	log.Printf("MockClient: GetUserCommentsURL generated: %s", url)
	return url
}

func (m *MockableRedditClient) GetPostURL(postID string) string {
	url := fmt.Sprintf("https://reddit.com/comments/%s.json?raw_json=1&sort=new", postID)
	log.Printf("MockClient: GetPostURL generated: %s", url)
	return url
}

func (m *MockableRedditClient) GetSearchURL(searchParams map[string]string) string {
	url := "https://reddit.com/search.json?raw_json=1"
	for key, value := range searchParams {
		url += fmt.Sprintf("&%s=%s", key, value)
	}
	log.Printf("MockClient: GetSearchURL generated: %s", url)
	return url
}

// Mock the config loading for integration tests
func mockConfig() *config.Config {
	return &config.Config{
		ProxyURLs:           []string{"http://mock-proxy.com"},
		UserAgent:           "Integration-Test-Agent",
		MaxRetries:          1,
		DefaultPostLimit:    10,
		DefaultCommentLimit: 10,
		ServerPort:          "8080",
		ReadTimeout:         5 * time.Second,
		WriteTimeout:        5 * time.Second,
		RedditBaseURL:       "https://reddit.com",
		RequestTimeout:      5 * time.Second,
		RateLimitDelay:      100 * time.Millisecond,
	}
}

// Create a test application with mocked client
func setupTestApp() (*echo.Echo, *MockableRedditClient) {
	// Create a mockable client that will intercept Reddit API calls
	mockClient := &MockableRedditClient{
		MockResponse: make(map[string]json.RawMessage),
	}
	
	// Create real parser
	redditParser := parser.NewRedditParser()
	
	// Create real scraper with mock client
	scraperService := scraper.NewScraperService(mockClient, redditParser)
	
	// Create Echo server
	e := echo.New()
	
	// Set up real routes with the scraper service
	router.NewRouter(e, scraperService)
	
	log.Println("Test app setup complete with mock client")
	return e, mockClient
}

// Test the integration between HTTP handlers, Scraper, and Parser
func TestSubredditEndpointIntegration(t *testing.T) {
	log.Println("======== Starting TestSubredditEndpointIntegration ========")
	
	// Setup test app
	e, mockClient := setupTestApp()
	
	// Partial URL to match
	subredditPartialURL := "/r/test/new.json"
	mockClient.MockResponse[subredditPartialURL] = json.RawMessage(`{
		"data": {
			"children": [
				{
					"kind": "t3",
					"data": {
						"id": "abc123",
						"title": "Integration Test Post",
						"selftext": "This is an integration test",
						"author": "tester",
						"score": 42,
						"created_utc": 1620000000,
						"subreddit": "test",
						"permalink": "/r/test/comments/abc123/integration_test",
						"url": "https://reddit.com/r/test/comments/abc123/integration_test"
					}
				}
			],
			"after": ""
		}
	}`)
	
	// Create request to subreddit endpoint
	req := httptest.NewRequest(http.MethodGet, "/subreddit?subreddit=test&limit=10", nil)
	rec := httptest.NewRecorder()
	
	log.Printf("Sending request to: %s", req.URL.String())
	
	// Handle the request using the Echo instance
	e.ServeHTTP(rec, req)
	
	// Check response
	log.Printf("Response status code: %d", rec.Code)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
		log.Printf("Response body: %s", rec.Body.String())
		t.FailNow()
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Log full response
	prettyResp, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("Response: %s", string(prettyResp))
	
	// Check posts in the response
	posts, ok := response["posts"].([]interface{})
	if !ok {
		t.Fatalf("Expected posts array in response, got %T", response["posts"])
	}
	
	log.Printf("Found %d posts in response", len(posts))
	
	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}
	
	// Verify first post properties
	if len(posts) > 0 {
		post := posts[0].(map[string]interface{})
		log.Printf("Post title: %s, author: %s", post["title"], post["author"])
		
		if post["title"] != "Integration Test Post" {
			t.Errorf("Expected post title 'Integration Test Post', got '%s'", post["title"])
		}
		if post["author"] != "tester" {
			t.Errorf("Expected post author 'tester', got '%s'", post["author"])
		}
	}
	
	log.Println("======== TestSubredditEndpointIntegration PASSED ========")
}

// Test the integration between HTTP handlers, Scraper, and Parser for User endpoint
func TestUserEndpointIntegration(t *testing.T) {
	log.Println("======== Starting TestUserEndpointIntegration ========")
	
	// Setup test app
	e, mockClient := setupTestApp()
	
	// Partial URLs to match
	userAboutPartialURL := "/user/tester/about.json"
	mockClient.MockResponse[userAboutPartialURL] = json.RawMessage(`{
		"data": {
			"name": "tester",
			"created_utc": 1620000000,
			"link_karma": 100,
			"comment_karma": 200
		}
	}`)
	
	userPostsPartialURL := "/user/tester/submitted/new.json"
	mockClient.MockResponse[userPostsPartialURL] = json.RawMessage(`{
		"data": {
			"children": [
				{
					"kind": "t3",
					"data": {
						"id": "post123",
						"title": "User Post",
						"selftext": "User post content",
						"author": "tester",
						"score": 10,
						"created_utc": 1620100000,
						"subreddit": "test",
						"permalink": "/r/test/comments/post123/user_post",
						"url": "https://reddit.com/r/test/comments/post123/user_post"
					}
				}
			],
			"after": ""
		}
	}`)
	
	userCommentsPartialURL := "/user/tester/comments/.json"
	mockClient.MockResponse[userCommentsPartialURL] = json.RawMessage(`{
		"data": {
			"children": [
				{
					"kind": "t1",
					"data": {
						"id": "comment123",
						"body": "User comment content",
						"author": "tester",
						"score": 5,
						"created_utc": 1620200000,
						"subreddit": "test",
						"link_id": "t3_post123",
						"link_title": "Some Post"
					}
				}
			],
			"after": ""
		}
	}`)
	
	// Create request to user endpoint
	req := httptest.NewRequest(http.MethodGet, "/user?username=tester&post_limit=10&comment_limit=10", nil)
	rec := httptest.NewRecorder()
	
	log.Printf("Sending request to: %s", req.URL.String())
	
	// Handle the request using the Echo instance
	e.ServeHTTP(rec, req)
	
	// Check response
	log.Printf("Response status code: %d", rec.Code)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
		log.Printf("Response body: %s", rec.Body.String())
		t.FailNow()
	}
	
	// Parse response
	var response models.UserActivity
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Log response details
	log.Printf("User info: %s, Link Karma: %d, Comment Karma: %d", 
		response.UserInfo.Username, response.UserInfo.LinkKarma, response.UserInfo.CommentKarma)
	log.Printf("Found %d posts and %d comments", len(response.Posts), len(response.Comments))
	
	// Check user info
	if response.UserInfo.Username != "tester" {
		t.Errorf("Expected username 'tester', got '%s'", response.UserInfo.Username)
	}
	
	// Check posts
	if len(response.Posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(response.Posts))
	} else {
		log.Printf("Post title: %s", response.Posts[0].Title)
		if response.Posts[0].Title != "User Post" {
			t.Errorf("Expected post title 'User Post', got '%s'", response.Posts[0].Title)
		}
	}
	
	// Check comments
	if len(response.Comments) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(response.Comments))
	} else {
		log.Printf("Comment body: %s", response.Comments[0].Body)
		if response.Comments[0].Body != "User comment content" {
			t.Errorf("Expected comment body 'User comment content', got '%s'", response.Comments[0].Body)
		}
	}
	
	log.Println("======== TestUserEndpointIntegration PASSED ========")
}

// Test the Post endpoint integration
func TestPostEndpointIntegration(t *testing.T) {
	log.Println("======== Starting TestPostEndpointIntegration ========")
	
	// Setup test app
	e, mockClient := setupTestApp()
	
	// Post ID to test
	postID := "abc123"
	
	// Partial URL to match
	postPartialURL := fmt.Sprintf("/comments/%s.json", postID)
	mockClient.MockResponse[postPartialURL] = json.RawMessage(`[
		{
			"data": {
				"children": [
					{
						"data": {
							"id": "abc123",
							"title": "Test Post",
							"author": "tester",
							"created_utc": 1620000000,
							"score": 42,
							"permalink": "/r/test/comments/abc123/test_post",
							"selftext": "This is a test post"
						}
					}
				]
			}
		},
		{
			"data": {
				"children": [
					{
						"kind": "t1",
						"data": {
							"id": "comment1",
							"author": "commenter",
							"body": "This is a comment",
							"score": 5,
							"created_utc": 1620000100,
							"replies": ""
						}
					}
				]
			}
		}
	]`)
	
	// Create request to post endpoint
	req := httptest.NewRequest(http.MethodGet, "/post?post_id="+postID, nil)
	rec := httptest.NewRecorder()
	
	log.Printf("Sending request to: %s", req.URL.String())
	
	// Handle the request using the Echo instance
	e.ServeHTTP(rec, req)
	
	// Check response
	log.Printf("Response status code: %d", rec.Code)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
		log.Printf("Response body: %s", rec.Body.String())
		t.FailNow()
	}
	
	// Parse response
	var postDetail models.PostDetail
	if err := json.Unmarshal(rec.Body.Bytes(), &postDetail); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Log post details
	log.Printf("Post ID: %s, Title: %s, Author: %s", 
		postDetail.Post.ID, postDetail.Post.Title, postDetail.Post.Author)
	log.Printf("Found %d comments", len(postDetail.Comments))
	
	// Check post details
	if postDetail.Post.ID != postID {
		t.Errorf("Expected post ID '%s', got '%s'", postID, postDetail.Post.ID)
	}
	
	if postDetail.Post.Title != "Test Post" {
		t.Errorf("Expected post title 'Test Post', got '%s'", postDetail.Post.Title)
	}
	
	// Check comments
	if len(postDetail.Comments) != 1 {
		t.Errorf("Expected 1 comment, got %d", len(postDetail.Comments))
	} else {
		log.Printf("Comment body: %s", postDetail.Comments[0].Body)
		if postDetail.Comments[0].Body != "This is a comment" {
			t.Errorf("Expected comment body 'This is a comment', got '%s'", postDetail.Comments[0].Body)
		}
	}
	
	log.Println("======== TestPostEndpointIntegration PASSED ========")
}

// Test the Search endpoint integration
func TestSearchEndpointIntegration(t *testing.T) {
	log.Println("======== Starting TestSearchEndpointIntegration ========")
	
	// Setup test app
	e, mockClient := setupTestApp()
	
	// Partial URL to match
	searchPartialURL := "/search.json"
	mockClient.MockResponse[searchPartialURL] = json.RawMessage(`{
		"data": {
			"children": [
				{
					"kind": "t3",
					"data": {
						"id": "search123",
						"title": "Search Result",
						"selftext": "This is a search result",
						"author": "searcher",
						"score": 15,
						"created_utc": 1620300000,
						"subreddit": "test",
						"permalink": "/r/test/comments/search123/search_result",
						"url": "https://reddit.com/r/test/comments/search123/search_result"
					}
				}
			],
			"after": ""
		}
	}`)
	
	// Create request to search endpoint
	req := httptest.NewRequest(http.MethodGet, "/search?search_string=test&limit=25", nil)
	rec := httptest.NewRecorder()
	
	log.Printf("Sending request to: %s", req.URL.String())
	
	// Handle the request using the Echo instance
	e.ServeHTTP(rec, req)
	
	// Check response
	log.Printf("Response status code: %d", rec.Code)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
		log.Printf("Response body: %s", rec.Body.String())
		t.FailNow()
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	
	// Log response details
	prettyResp, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("Response: %s", string(prettyResp))
	
	// Check posts in the response
	posts, ok := response["posts"].([]interface{})
	if !ok {
		t.Fatalf("Expected posts array in response, got %T", response["posts"])
	}
	
	log.Printf("Found %d search results", len(posts))
	
	if len(posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(posts))
	}
	
	// Verify first search result properties
	if len(posts) > 0 {
		post := posts[0].(map[string]interface{})
		log.Printf("Search result title: %s, author: %s", post["title"], post["author"])
		
		if post["title"] != "Search Result" {
			t.Errorf("Expected post title 'Search Result', got '%s'", post["title"])
		}
		if post["author"] != "searcher" {
			t.Errorf("Expected post author 'searcher', got '%s'", post["author"])
		}
	}
	
	log.Println("======== TestSearchEndpointIntegration PASSED ========")
}