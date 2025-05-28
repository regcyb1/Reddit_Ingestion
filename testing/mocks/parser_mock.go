package mocks

import (
	"context"
	"encoding/json"
	
	"reddit-ingestion/internal/models"
)

type MockParser struct {
	ParseSubredditFunc     func(ctx context.Context, data json.RawMessage) ([]models.Post, string, error)
	ParseUserInfoFunc      func(ctx context.Context, data json.RawMessage) (models.UserInfo, error)
	ParseUserPostsFunc     func(ctx context.Context, data json.RawMessage) ([]models.UserPost, string, error)
	ParseUserCommentsFunc  func(ctx context.Context, data json.RawMessage) ([]models.UserComment, string, error)
	ParsePostFunc          func(ctx context.Context, postData, commentData json.RawMessage) (models.PostDetail, error)
	ParseMoreCommentsFunc  func(ctx context.Context, data json.RawMessage) ([]models.Comment, error)
}

func (m *MockParser) ParseSubreddit(ctx context.Context, data json.RawMessage) ([]models.Post, string, error) {
	return m.ParseSubredditFunc(ctx, data)
}

func (m *MockParser) ParseUserInfo(ctx context.Context, data json.RawMessage) (models.UserInfo, error) {
	return m.ParseUserInfoFunc(ctx, data)
}

func (m *MockParser) ParseUserPosts(ctx context.Context, data json.RawMessage) ([]models.UserPost, string, error) {
	return m.ParseUserPostsFunc(ctx, data)
}

func (m *MockParser) ParseUserComments(ctx context.Context, data json.RawMessage) ([]models.UserComment, string, error) {
	return m.ParseUserCommentsFunc(ctx, data)
}

func (m *MockParser) ParsePost(ctx context.Context, postData, commentData json.RawMessage) (models.PostDetail, error) {
	return m.ParsePostFunc(ctx, postData, commentData)
}

func (m *MockParser) ParseMoreComments(ctx context.Context, data json.RawMessage) ([]models.Comment, error) {
	return m.ParseMoreCommentsFunc(ctx, data)
}
