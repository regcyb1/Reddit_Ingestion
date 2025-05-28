// internal/parser/parser.go
package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"reddit-ingestion/internal/models"

	"github.com/google/uuid"
)

// Parser defines the interface for parsing Reddit API responses
type Parser interface {
	ParseSubreddit(ctx context.Context, data json.RawMessage) ([]models.Post, string, error)
	ParseUserInfo(ctx context.Context, data json.RawMessage) (models.UserInfo, error)
	ParseUserPosts(ctx context.Context, data json.RawMessage) ([]models.UserPost, string, error)
	ParseUserComments(ctx context.Context, data json.RawMessage) ([]models.UserComment, string, error)
	ParsePost(ctx context.Context, postData, commentData json.RawMessage) (models.PostDetail, error)
	ParseMoreComments(ctx context.Context, data json.RawMessage) ([]models.Comment, error)
}

type RedditParser struct{}

func NewRedditParser() *RedditParser {
	return &RedditParser{}
}

func (p *RedditParser) ParseSubreddit(ctx context.Context, data json.RawMessage) ([]models.Post, string, error) {
	var listing struct {
		Data struct {
			Children []struct {
				Kind string `json:"kind"`
				Data struct {
					ID            string  `json:"id"`
					Title         string  `json:"title"`
					Selftext      string  `json:"selftext"`
					Author        string  `json:"author"`
					Score         int     `json:"score"`
					CreatedUTC    float64 `json:"created_utc"`
					Subreddit     string  `json:"subreddit"`
					LinkFlairText string  `json:"link_flair_text"`
					Permalink     string  `json:"permalink"`
					URL           string  `json:"url"`
				} `json:"data"`
			} `json:"children"`
			After string `json:"after"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, "", fmt.Errorf("parse subreddit JSON: %w", err)
	}

	var posts []models.Post
	for _, child := range listing.Data.Children {
		if child.Kind != "t3" {
			continue
		}

		created := time.Unix(int64(child.Data.CreatedUTC), 0)

		posts = append(posts, models.Post{
			ID:        child.Data.ID,
			Title:     child.Data.Title,
			Body:      child.Data.Selftext,
			Author:    child.Data.Author,
			Score:     child.Data.Score,
			CreatedAt: created,
			Flair:     child.Data.LinkFlairText,
			URL:       "https://reddit.com" + child.Data.Permalink,
		})
	}

	return posts, listing.Data.After, nil
}

func (p *RedditParser) ParseUserInfo(ctx context.Context, data json.RawMessage) (models.UserInfo, error) {
	var about struct {
		Data struct {
			Name         string  `json:"name"`
			CreatedUTC   float64 `json:"created_utc"`
			LinkKarma    int     `json:"link_karma"`
			CommentKarma int     `json:"comment_karma"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &about); err != nil {
		return models.UserInfo{}, fmt.Errorf("parse user info JSON: %w", err)
	}

	return models.UserInfo{
		Username:     about.Data.Name,
		LinkKarma:    about.Data.LinkKarma,
		CommentKarma: about.Data.CommentKarma,
		CreatedAt:    time.Unix(int64(about.Data.CreatedUTC), 0),
	}, nil
}

func (p *RedditParser) ParseUserPosts(ctx context.Context, data json.RawMessage) ([]models.UserPost, string, error) {
	var listing struct {
		Data struct {
			Children []struct {
				Kind string `json:"kind"`
				Data struct {
					ID            string  `json:"id"`
					Title         string  `json:"title"`
					Selftext      string  `json:"selftext"`
					Author        string  `json:"author"`
					Score         int     `json:"score"`
					CreatedUTC    float64 `json:"created_utc"`
					Subreddit     string  `json:"subreddit"`
					LinkFlairText string  `json:"link_flair_text"`
					Permalink     string  `json:"permalink"`
					URL           string  `json:"url"`
				} `json:"data"`
			} `json:"children"`
			After string `json:"after"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, "", fmt.Errorf("parse user posts JSON: %w", err)
	}

	var posts []models.UserPost
	for _, child := range listing.Data.Children {
		if child.Kind != "t3" {
			continue
		}

		created := time.Unix(int64(child.Data.CreatedUTC), 0)

		posts = append(posts, models.UserPost{
			ID:        child.Data.ID,
			Title:     child.Data.Title,
			Body:      child.Data.Selftext,
			Score:     child.Data.Score,
			CreatedAt: created,
			Subreddit: child.Data.Subreddit,
			Flair:     child.Data.LinkFlairText,
			URL:       "https://reddit.com" + child.Data.Permalink,
		})
	}

	return posts, listing.Data.After, nil
}

func (p *RedditParser) ParseUserComments(ctx context.Context, data json.RawMessage) ([]models.UserComment, string, error) {
	var listing struct {
		Data struct {
			Children []struct {
				Kind string `json:"kind"`
				Data struct {
					ID         string  `json:"id"`
					Body       string  `json:"body"`
					Author     string  `json:"author"`
					Score      int     `json:"score"`
					CreatedUTC float64 `json:"created_utc"`
					Subreddit  string  `json:"subreddit"`
					LinkID     string  `json:"link_id"`
					LinkTitle  string  `json:"link_title"`
					ParentID   string  `json:"parent_id"`
				} `json:"data"`
			} `json:"children"`
			After string `json:"after"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &listing); err != nil {
		return nil, "", fmt.Errorf("parse user comments JSON: %w", err)
	}

	var comments []models.UserComment
	for _, child := range listing.Data.Children {
		if child.Kind != "t1" {
			continue
		}

		created := time.Unix(int64(child.Data.CreatedUTC), 0)
		postID := child.Data.LinkID
		if len(postID) > 3 {
			postID = postID[3:] // Remove "t3_" prefix
		}

		comments = append(comments, models.UserComment{
			ID:        child.Data.ID,
			Body:      child.Data.Body,
			Score:     child.Data.Score,
			CreatedAt: created,
			Subreddit: child.Data.Subreddit,
			PostID:    postID,
			PostTitle: child.Data.LinkTitle,
		})
	}

	return comments, listing.Data.After, nil
}

func (p *RedditParser) ParsePost(ctx context.Context, postData, commentData json.RawMessage) (models.PostDetail, error) {
	var postBlock struct {
		Data struct {
			Children []struct {
				Data struct {
					ID            string  `json:"id"`
					Title         string  `json:"title"`
					Author        string  `json:"author"`
					CreatedUTC    float64 `json:"created_utc"`
					Score         int     `json:"score"`
					LinkFlairText string  `json:"link_flair_text"`
					Permalink     string  `json:"permalink"`
					Selftext      string  `json:"selftext"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}

	if err := json.Unmarshal(postData, &postBlock); err != nil {
		return models.PostDetail{}, fmt.Errorf("parse post JSON: %w", err)
	}

	if len(postBlock.Data.Children) == 0 {
		return models.PostDetail{}, fmt.Errorf("post not found")
	}

	pd := postBlock.Data.Children[0].Data
	post := models.Post{
		ID:        pd.ID,
		Title:     pd.Title,
		Body:      pd.Selftext,
		Author:    pd.Author,
		Score:     pd.Score,
		CreatedAt: time.Unix(int64(pd.CreatedUTC), 0),
		Flair:     pd.LinkFlairText,
		URL:       "https://old.reddit.com" + pd.Permalink,
	}

	comments, err := p.parseCommentsTree(ctx, commentData)
	if err != nil {
		return models.PostDetail{Post: post}, fmt.Errorf("parse comments: %w", err)
	}

	return models.PostDetail{Post: post, Comments: comments}, nil
}

func (p *RedditParser) ParseMoreComments(ctx context.Context, data json.RawMessage) ([]models.Comment, error) {
    var wrapper struct {
        JSON struct {
            Data struct {
                Things []models.RawChild `json:"things"`
            } `json:"data"`
        } `json:"json"`
    }
    
    if err := json.Unmarshal(data, &wrapper); err != nil {
        var directThings []models.RawChild
        if err2 := json.Unmarshal(data, &directThings); err2 == nil {
            return p.processComments(ctx, directThings), nil
        }
        
        return nil, fmt.Errorf("parse more comments JSON: %w", err)
    }
    
    thingCount := len(wrapper.JSON.Data.Things)
    fmt.Printf("Received %d things in morecomments response\n", thingCount)
    
    for i, thing := range wrapper.JSON.Data.Things {
        if i < 3 {
            fmt.Printf("Thing %d: Kind=%s, ID=%s, Author=%s\n", 
                i, thing.Kind, thing.Data.ID, thing.Data.Author)
        }
    }
    
    processed := p.processComments(ctx, wrapper.JSON.Data.Things)
    fmt.Printf("Processed %d comments from morecomments response\n", len(processed))
    return processed, nil
}

func (p *RedditParser) parseCommentsTree(ctx context.Context, data json.RawMessage) ([]models.Comment, error) {
	var commentsBlock struct {
		Data struct {
			Children []models.RawChild `json:"children"`
		} `json:"data"`
	}

	if err := json.Unmarshal(data, &commentsBlock); err != nil {
		return nil, fmt.Errorf("parse comments JSON: %w", err)
	}

	return p.processComments(ctx, commentsBlock.Data.Children), nil
}

func (p *RedditParser) processComments(ctx context.Context, children []models.RawChild) []models.Comment {
    var comments []models.Comment
    
    for _, child := range children {
        if ctx.Err() != nil {
            return comments
        }
        
        switch child.Kind {
        case "t1": // Regular comment
            comment := models.Comment{
                ID:        child.Data.ID,
                Author:    child.Data.Author,
                Body:      child.Data.Body,
                Score:     child.Data.Score,
                CreatedAt: time.Unix(int64(child.Data.CreatedUTC), 0),
            }
            
            // Process replies if they exist
            if len(child.Data.Replies) > 0 {
                var replies struct {
                    Data struct {
                        Children []models.RawChild `json:"children"`
                    } `json:"data"`
                }
                
                if err := json.Unmarshal(child.Data.Replies, &replies); err == nil {
                    comment.Replies = p.processComments(ctx, replies.Data.Children)
                    
                    // Check for "more" comments
                    for _, replyChild := range replies.Data.Children {
                        if replyChild.Kind == "more" && len(replyChild.Data.Children) > 0 {
                            comment.HasMore = true
                            comment.MoreIDs = append(comment.MoreIDs, replyChild.Data.Children...)
                        }
                    }
                }
            }
            
            comments = append(comments, comment)
            
        case "more": // "Load more comments" link
            if len(child.Data.Children) > 0 {
                // Check for "continue thread" links
                shouldSkip := false
                for _, id := range child.Data.Children {
                    if id == "continue" {
                        shouldSkip = true
                        fmt.Printf("Found 'continue' link in more comments, marked for special handling\n")
                        break
                    }
                }
                
                if !shouldSkip {
                    // Regular "more comments"
                    moreComment := models.Comment{
                        ID:      "more_" + uuid.New().String(),
                        IsMore:  true,
                        MoreIDs: child.Data.Children,
                    }
                    
                    fmt.Printf("Found 'more' comment with %d child IDs\n", len(child.Data.Children))
                    comments = append(comments, moreComment)
                } else {
                    // Add the "continue" as a special type
                    continueComment := models.Comment{
                        ID:       "continue_" + uuid.New().String(),
                        IsMore:   true,         // Still mark as "more" for compatibility
                        MoreIDs:  []string{child.Data.ParentID}, // Store parent ID
                        HasMore:  true,         // Use HasMore flag for "continue" links
                    }
                    comments = append(comments, continueComment)
                    fmt.Printf("Added 'continue' link as special comment type\n")
                }
            }
        }
    }
    
    return comments
}