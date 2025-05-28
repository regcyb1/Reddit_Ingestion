// internal/client/interface.go
package client

import (
	"context"
	"encoding/json"
)

type RedditClientInterface interface {
	FetchJSON(ctx context.Context, url string) (json.RawMessage, error)
	FetchMoreComments(ctx context.Context, postID string, commentIDs []string) (json.RawMessage, error)
	GetSubredditURL(subreddit string, limit int, after string) string
	GetUserAboutURL(username string) string
	GetUserPostsURL(username string, after string) string
	GetUserCommentsURL(username string, after string) string
	GetPostURL(postID string) string
	GetSearchURL(searchParams map[string]string) string
}