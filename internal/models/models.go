package models

import (
	"encoding/json"
	"time"
)

// Post represents a Reddit post
// swagger:model Post
type Post struct {
	// Reddit post ID
	ID string `json:"id"`
	// Post title
	Title string `json:"title"`
	// Post body/content
	Body string `json:"body"`
	// Author's username
	Author string `json:"author"`
	// Post score (upvotes minus downvotes)
	Score int `json:"score"`
	// Creation timestamp
	CreatedAt time.Time `json:"created_at"`
	// Post flair text
	Flair string `json:"flair,omitempty"`
	// Full URL to the post
	URL string `json:"url"`
}

// Comment represents a Reddit comment
// swagger:model Comment
type Comment struct {
	// Comment ID
	ID string `json:"id"`
	// Comment author's username
	Author string `json:"author"`
	// Comment body text
	Body string `json:"body"`
	// Comment score
	Score int `json:"score"`
	// Comment creation timestamp
	CreatedAt time.Time `json:"created_at"`
	// Nested comment replies
	Replies []Comment `json:"replies,omitempty"`
	// Flag indicating if this is a "more comments" placeholder
	IsMore bool `json:"is_more,omitempty"`
	// List of IDs for additional comments that need to be loaded
    MoreIDs []string `json:"more_ids,omitempty"`
    // Flag indicating if there are more child comments available
    HasMore bool `json:"has_more,omitempty"`
	// Count of total remaining comments in a "more" object
    MoreCount int `json:"more_count,omitempty"`

}

// UserInfo represents a Reddit user's profile information
// swagger:model UserInfo
type UserInfo struct {
	// Username
	Username string `json:"username"`
	// Link karma score
	LinkKarma int `json:"link_karma"`
	// Comment karma score
	CommentKarma int `json:"comment_karma"`
	// Account creation timestamp
	CreatedAt time.Time `json:"created_at"`
}

// PostDetail represents a Reddit post with its comments
// swagger:model PostDetail
type PostDetail struct {
	// Post information
	Post Post `json:"post"`
	// Comments on the post
	Comments []Comment `json:"comments"`
}
// UserComment represents a comment made by a user
// swagger:model UserComment
type UserComment struct {
	// Comment ID
	ID string `json:"id"`
	// Comment body text
	Body string `json:"body"`
	// Comment score
	Score int `json:"score"`
	// Comment creation timestamp
	CreatedAt time.Time `json:"created_at"`
	// Subreddit where the comment was posted
	Subreddit string `json:"subreddit"`
	// ID of the post containing this comment
	PostID string `json:"post_id"`
	// Title of the post containing this comment
	PostTitle string `json:"post_title"`
	// Author of the parent comment (if this is a reply)
	ParentAuthor string `json:"parent_author,omitempty"`
}

// UserPost represents a post made by a user
// swagger:model UserPost
type UserPost struct {
	// Post ID
	ID string `json:"id"`
	// Post title
	Title string `json:"title"`
	// Post body/content
	Body string `json:"body"`
	// Post score
	Score int `json:"score"`
	// Post creation timestamp
	CreatedAt time.Time `json:"created_at"`
	// Subreddit where the post was created
	Subreddit string `json:"subreddit"`
	// Full URL to the post
	URL string `json:"url"`
	// Post flair text
	Flair string `json:"flair,omitempty"`
}

// UserActivity represents all activity for a specific user
// swagger:model UserActivity
type UserActivity struct {
	// User profile information
	UserInfo UserInfo `json:"user_info"`
	// Posts created by the user
	Posts []UserPost `json:"posts,omitempty"`
	// Comments made by the user
	Comments []UserComment `json:"comments,omitempty"`
}

// RawChild is an internal structure used for parsing Reddit API responses
type RawChild struct {
	Kind string `json:"kind"`
	Data struct {
		ID string `json:"id"`
		Author string `json:"author"`
		Body string `json:"body"`
		Score int `json:"score"`
		CreatedUTC float64 `json:"created_utc"`
		Replies json.RawMessage `json:"replies"`
		Children []string `json:"children"`
		ParentID string `json:"parent_id"`
		Count int `json:"count"`
		Permalink string `json:"permalink"`
	} `json:"data"`
}