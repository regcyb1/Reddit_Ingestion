package mocks

import (
	"context"
	"encoding/json"
)

type MockRedditClient struct {
	FetchJSONFunc          func(ctx context.Context, url string) (json.RawMessage, error)
	FetchMoreCommentsFunc  func(ctx context.Context, postID string, commentIDs []string) (json.RawMessage, error)
	GetSubredditURLFunc    func(subreddit string, limit int, after string) string
	GetUserAboutURLFunc    func(username string) string
	GetUserPostsURLFunc    func(username string, after string) string
	GetUserCommentsURLFunc func(username string, after string) string
	GetPostURLFunc         func(postID string) string
	GetSearchURLFunc       func(searchParams map[string]string) string
}

func (m *MockRedditClient) FetchJSON(ctx context.Context, url string) (json.RawMessage, error) {
	return m.FetchJSONFunc(ctx, url)
}

func (m *MockRedditClient) FetchMoreComments(ctx context.Context, postID string, commentIDs []string) (json.RawMessage, error) {
	return m.FetchMoreCommentsFunc(ctx, postID, commentIDs)
}

func (m *MockRedditClient) GetSubredditURL(subreddit string, limit int, after string) string {
	return m.GetSubredditURLFunc(subreddit, limit, after)
}

func (m *MockRedditClient) GetUserAboutURL(username string) string {
	return m.GetUserAboutURLFunc(username)
}

func (m *MockRedditClient) GetUserPostsURL(username string, after string) string {
	return m.GetUserPostsURLFunc(username, after)
}

func (m *MockRedditClient) GetUserCommentsURL(username string, after string) string {
	return m.GetUserCommentsURLFunc(username, after)
}

func (m *MockRedditClient) GetPostURL(postID string) string {
	return m.GetPostURLFunc(postID)
}

func (m *MockRedditClient) GetSearchURL(searchParams map[string]string) string {
	return m.GetSearchURLFunc(searchParams)
}
