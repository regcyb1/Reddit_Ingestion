package models

// ErrorResponse represents an error response
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message
	Message string `json:"message"`
}

// SubredditResponse represents a response for the subreddit endpoint
// swagger:model SubredditResponse
type SubredditResponse struct {
	// List of posts
	Posts []Post `json:"posts"`
	// Metadata about the request
	Meta struct {
		// Requested limit
		RequestedLimit int `json:"requested_limit"`
		// Actual count of posts returned
		ActualCount int `json:"actual_count"`
		// Subreddit name
		Subreddit string `json:"subreddit"`
		// Since timestamp (Unix time)
		SinceTimestamp int64 `json:"since_timestamp"`
		// Processing time in milliseconds
		ProcessingTimeMs int64 `json:"processing_time_ms"`
	} `json:"meta"`
}

// SearchResponse represents a response for the search endpoint
// swagger:model SearchResponse
type SearchResponse struct {
	// List of posts matching the search
	Posts []Post `json:"posts"`
	// Metadata about the search
	Meta struct {
		// Search query
		Query string `json:"query"`
		// Search parameters used
		Params map[string]string `json:"params"`
		// Count of posts returned
		Count int `json:"count"`
		// Processing time in milliseconds
		ProcessingTimeMs int64 `json:"processing_time_ms"`
		// Requested limit description
		RequestedLimit string `json:"requested_limit"`
	} `json:"meta"`
}