// internal/client/reddit_client.go
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"reddit-ingestion/internal/config"
	"reddit-ingestion/pkg/utils"
)

type RedditClient struct {
	client     *utils.RetryableClient
	userAgent  string
	config     *config.Config
	baseURL    string
}

func NewRedditClient(cfg *config.Config) (*RedditClient, error) {
	if cfg.UserAgent == "" {
		return nil, fmt.Errorf("REDDIT_USER_AGENT environment variable is required")
	}
	
	if len(cfg.ProxyURLs) == 0 {
		return nil, fmt.Errorf("at least one proxy URL must be provided")
	}
	
	fmt.Printf("Initializing Reddit client with %d proxies\n", len(cfg.ProxyURLs))
	
	for i, proxy := range cfg.ProxyURLs {
		maskedProxy := maskProxyURL(proxy)
		fmt.Printf("Proxy #%d: %s\n", i+1, maskedProxy)
	}

	client, err := utils.NewRetryableClient(
		cfg.ProxyURLs,
		cfg.MaxRetries,
		cfg.UserAgent,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}
	
	return &RedditClient{
		client:    client,
		userAgent: cfg.UserAgent,
		config:    cfg,
		baseURL:   cfg.RedditBaseURL,
	}, nil
}

func maskProxyURL(proxyURL string) string {
	if !strings.Contains(proxyURL, "@") {
		return proxyURL
	}
	
	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		parts := strings.Split(proxyURL, "@")
		if len(parts) > 1 {
			auth := strings.Split(parts[0], "://")
			protocol := ""
			if len(auth) > 1 {
				protocol = auth[0] + "://"
				auth[0] = auth[1]
			}
			
			userPass := strings.Split(auth[0], ":")
			if len(userPass) > 1 {
				return protocol + userPass[0] + ":****@" + parts[1]
			}
		}
		return "[masked]"
	}
	
	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		return strings.Replace(proxyURL, parsedURL.User.String(), username+":****", 1)
	}
	
	return proxyURL
}

func (r *RedditClient) FetchJSON(ctx context.Context, url string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	_, bodyBytes, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetchJSON request: %w", err)
	}
	
	return bodyBytes, nil
}

func (r *RedditClient) GetSubredditURL(subreddit string, limit int, after string) string {
	baseURL := fmt.Sprintf("%s/r/%s/new.json?raw_json=1", r.baseURL, subreddit)
	
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if after != "" {
		params.Set("after", after)
	}
	
	paramsStr := params.Encode()
	if paramsStr != "" {
		baseURL += "&" + paramsStr
	}
	
	return baseURL
}

func (r *RedditClient) GetUserAboutURL(username string) string {
	return fmt.Sprintf("%s/user/%s/about.json", r.baseURL, username)
}

func (r *RedditClient) GetUserPostsURL(username string, after string) string {
	baseURL := fmt.Sprintf("%s/user/%s/submitted/new.json?raw_json=1&sort=new", r.baseURL, username)
	
	if after != "" {
		baseURL += "&after=" + after
	}
	
	return baseURL
}

func (r *RedditClient) GetUserCommentsURL(username string, after string) string {
	baseURL := fmt.Sprintf("%s/user/%s/comments/.json?raw_json=1&limit=100", r.baseURL, username)
	
	if after != "" {
		baseURL += "&after=" + after
	}
	
	return baseURL
}

func (r *RedditClient) GetPostURL(postID string) string {
	return fmt.Sprintf("%s/comments/%s.json?raw_json=1&sort=new", r.baseURL, postID)
}

func (r *RedditClient) FetchMoreComments(ctx context.Context, postID string, commentIDs []string) (json.RawMessage, error) {
    if len(commentIDs) == 0 {
        return nil, nil
    }
    
    fullPostID := postID
    if !strings.HasPrefix(fullPostID, "t3_") {
        fullPostID = "t3_" + postID
    }
    
    endpoint := "https://api.reddit.com/api/morechildren"
    
    params := url.Values{
        "api_type":       {"json"},
        "link_id":        {fullPostID},
        "children":       {strings.Join(commentIDs, ",")},
        "limit_children": {"false"},
        "sort":           {"new"},
    }
    
    // Log the request
    fmt.Printf("Fetching %d more comments for post %s\n", len(commentIDs), postID)
    
    // Add retry logic
    maxRetries := 3
    var lastErr error
    
    for retry := 0; retry < maxRetries; retry++ {
        if retry > 0 {
            // Exponential backoff
            waitTime := time.Duration(math.Pow(2, float64(retry))) * time.Second
            fmt.Printf("Retrying morechildren request after %v (attempt %d/%d)\n", 
                waitTime, retry+1, maxRetries)
            time.Sleep(waitTime)
        }
        
        req, err := http.NewRequestWithContext(ctx, "GET", endpoint+"?"+params.Encode(), nil)
        if err != nil {
            lastErr = fmt.Errorf("create request: %w", err)
            continue
        }
        
        req.Header.Set("User-Agent", r.userAgent)
        
        _, bodyBytes, err := r.client.Do(req)
        if err == nil {
            return bodyBytes, nil
        }
        
        lastErr = fmt.Errorf("fetchMoreComments request: %w", err)
        
        // Check if rate limited
        if strings.Contains(err.Error(), "429") {
            fmt.Println("Rate limited by Reddit API, waiting longer...")
            time.Sleep(30 * time.Second)
        }
    }
    
    return nil, fmt.Errorf("max retries exceeded: %w, For comments: %v", lastErr, commentIDs)
}

func (r *RedditClient) GetSearchURL(searchParams map[string]string) string {
	baseSearchURL := fmt.Sprintf("%s/search.json?raw_json=1", r.baseURL)
	
	params := url.Values{}
	
	var queryParts []string
	
	if search, ok := searchParams["search_string"]; ok && search != "" {
		queryParts = append(queryParts, search)
	}
	
	advancedParams := []string{"subreddit", "author", "site", "url", "selftext", "self", "nsfw"}
	for _, param := range advancedParams {
		if value, ok := searchParams[param]; ok && value != "" {
			queryParts = append(queryParts, fmt.Sprintf("%s:%s", param, value))
		}
	}
	
	if len(queryParts) > 0 {
		params.Set("q", strings.Join(queryParts, " "))
	}
	
	directParams := []string{"sort", "t", "limit", "after", "before", "restrict_sr"}
	for _, param := range directParams {
		if value, ok := searchParams[param]; ok && value != "" {
			params.Set(param, value)
		}
	}
	
	if time, ok := searchParams["time"]; ok && time != "" {
		params.Set("t", time)
	}
	
	paramsStr := params.Encode()
	if paramsStr != "" {
		baseSearchURL += "&" + paramsStr
	}
	
	return baseSearchURL
}