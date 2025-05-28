// internal/parser/interface.go
package parser

import (
	"context"
	"encoding/json"
	
	"reddit-ingestion/internal/models"
)

type ParserInterface interface {
	ParseSubreddit(ctx context.Context, data json.RawMessage) ([]models.Post, string, error)
	ParseUserInfo(ctx context.Context, data json.RawMessage) (models.UserInfo, error)
	ParseUserPosts(ctx context.Context, data json.RawMessage) ([]models.UserPost, string, error)
	ParseUserComments(ctx context.Context, data json.RawMessage) ([]models.UserComment, string, error)
	ParsePost(ctx context.Context, postData, commentData json.RawMessage) (models.PostDetail, error)
	ParseMoreComments(ctx context.Context, data json.RawMessage) ([]models.Comment, error)
}