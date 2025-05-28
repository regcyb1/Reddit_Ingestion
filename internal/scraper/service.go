// internal/scraper/service.go
package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"reddit-ingestion/internal/client"
	"reddit-ingestion/internal/models"
	"reddit-ingestion/internal/parser"
)

// ScraperService defines the interface for scraping Reddit content
type ScraperService interface {
	ScrapeSubreddit(ctx context.Context, subreddit string, sinceTimestamp int64, limit int) ([]models.Post, error)
	ScrapeUserActivity(ctx context.Context, username string, sinceTimestamp int64, postLimit, commentLimit int) (models.UserActivity, error)
	ScrapePost(ctx context.Context, postID string) (models.PostDetail, error)
	Search(ctx context.Context, searchParams map[string]string, sinceTimestamp int64, limit int) ([]models.Post, error)
}

type scraperService struct {
	client client.RedditClientInterface
	parser parser.ParserInterface
}

type MoreCommentSet struct {
    Parent        string   
    CommentIDs    []string 
    Depth         int      
    PlaceholderID string   
}

func NewScraperService(client client.RedditClientInterface, parser parser.ParserInterface) ScraperService {
	return &scraperService{
		client: client,
		parser: parser,
	}
}

// ScrapeSubreddit retrieves posts from a subreddit
func (s *scraperService) ScrapeSubreddit(
	ctx context.Context,
	subreddit string,
	sinceTimestamp int64,
	limit int,
) ([]models.Post, error) {
	startTime := time.Now()
	var posts []models.Post

	// Case 1: No timestamp and limit 0 - fetch only first page with default size
	if sinceTimestamp == 0 && limit == 0 {
		fmt.Printf("No timestamp or limit provided, fetching only the first page for subreddit %s\n", subreddit)

		apiURL := s.client.GetSubredditURL(subreddit, 0, "")

		data, err := s.client.FetchJSON(ctx, apiURL)
		if err != nil {
			return nil, fmt.Errorf("fetch subreddit: %w", err)
		}

		pagePosts, _, err := s.parser.ParseSubreddit(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("parse subreddit: %w", err)
		}

		posts = append(posts, pagePosts...)

		fmt.Printf("First page fetch yielded %d posts\n", len(posts))
		fmt.Printf("Final result: %d posts fetched in %v\n", len(posts), time.Since(startTime))
		return posts, nil
	}

	apiLimit := 100 // Maximum allowed by Reddit API per page
	
	// Special case: if limit is -1, we retrieve all posts (no post limit)
	if limit > 0 && limit < apiLimit {
		apiLimit = limit
	}

	after := ""
	pageCount := 0
	maxPages := 20 

	// Special case: if limit is -1, set a very high max pages value
	if limit == -1 {
		maxPages = 1000
		fmt.Printf("Special case: limit = -1, attempting to scrape ALL posts from subreddit %s\n", subreddit)
	}
	
	// Increase max pages for timestamp filtering
	if sinceTimestamp > 0 {
		if limit != -1 {
			maxPages = 20
		}
	}

	for pageCount < maxPages {
		if ctx.Err() != nil {
			return posts, ctx.Err()
		}

		pageCount++

		apiURL := s.client.GetSubredditURL(subreddit, apiLimit, after)
		fmt.Printf("Fetching page %d for subreddit %s (URL: %s)\n", pageCount, subreddit, apiURL)

		data, err := s.client.FetchJSON(ctx, apiURL)
		if err != nil {
			return nil, fmt.Errorf("fetch subreddit: %w", err)
		}

		pagePosts, nextAfter, err := s.parser.ParseSubreddit(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("parse subreddit: %w", err)
		}

		pagePostCount := 0
		reachedTimeLimit := false

		// Filter by timestamp if needed
		for _, post := range pagePosts {
			if sinceTimestamp > 0 && post.CreatedAt.Unix() < sinceTimestamp {
				reachedTimeLimit = true
				continue
			}

			pagePostCount++
			posts = append(posts, post)
		}

		fmt.Printf("Page %d yielded %d posts (total now: %d/%d)\n",
			pageCount, pagePostCount, len(posts), limit)

		// Stop conditions
		if limit > 0 && len(posts) >= limit {
			fmt.Println("Reached requested limit, stopping pagination")
			break
		}

		if reachedTimeLimit {
			fmt.Println("Reached time limit cutoff, stopping pagination")
			break
		}

		if nextAfter == "" || pagePostCount == 0 {
			fmt.Println("No more pages available or empty page")
			break
		}

		after = nextAfter

		// Timeout handling
		timeoutDuration := 30 * time.Second
		if limit == -1 {
			timeoutDuration = 3 * time.Minute
		}
		
		if time.Since(startTime) > timeoutDuration && len(posts) > 0 {
			if limit == -1 {
				fmt.Printf("Extended time limit (%v) for full scraping reached, returning results so far\n", timeoutDuration)
			} else {
				fmt.Printf("Time limit (%v) for request reached, returning results so far\n", timeoutDuration)
			}
			break
		}
	}

	// Apply limit if necessary, but not when limit is -1
	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	fmt.Printf("Final result: %d posts fetched in %v\n", len(posts), time.Since(startTime))
	return posts, nil
}

// ScrapeUserActivity retrieves a user's activity on Reddit
func (s *scraperService) ScrapeUserActivity(
	ctx context.Context,
	username string,
	sinceTimestamp int64,
	postLimit, commentLimit int,
) (models.UserActivity, error) {
	activity := models.UserActivity{}

	aboutURL := s.client.GetUserAboutURL(username)

	aboutData, err := s.client.FetchJSON(ctx, aboutURL)
	if err != nil {
		return activity, fmt.Errorf("fetch user info: %w", err)
	}

	userInfo, err := s.parser.ParseUserInfo(ctx, aboutData)
	if err != nil {
		return activity, fmt.Errorf("parse user info: %w", err)
	}

	activity.UserInfo = userInfo

	var wg sync.WaitGroup
	var postsErr, commentsErr error
	postsChan := make(chan []models.UserPost, 1)
	commentsChan := make(chan []models.UserComment, 1)

	wg.Add(2)

	// Fetch posts concurrently
	go func() {
		defer wg.Done()
		posts, err := s.fetchUserPosts(ctx, username, sinceTimestamp, postLimit)
		if err != nil {
			postsErr = fmt.Errorf("fetch user posts: %w", err)
			return
		}
		postsChan <- posts
	}()

	// Fetch comments concurrently
	go func() {
		defer wg.Done()
		comments, err := s.fetchUserComments(ctx, username, sinceTimestamp, commentLimit)
		if err != nil {
			commentsErr = fmt.Errorf("fetch user comments: %w", err)
			return
		}
		commentsChan <- comments
	}()

	wg.Wait()
	close(postsChan)
	close(commentsChan)

	// Check for errors
	if postsErr != nil {
		return activity, postsErr
	}
	if commentsErr != nil {
		return activity, commentsErr
	}

	if posts, ok := <-postsChan; ok {
		activity.Posts = posts
	}

	if comments, ok := <-commentsChan; ok {
		activity.Comments = comments
	}

	return activity, nil
}

// fetchUserPosts function
func (s *scraperService) fetchUserPosts(
	ctx context.Context,
	username string,
	sinceTimestamp int64,
	limit int,
) ([]models.UserPost, error) {
	var posts []models.UserPost
	after := ""
	pageCount := 0
	startTime := time.Now()

	var needMultiplePages bool
	var maxPages int
	var effectiveLimit int

	switch {
	case limit == 0:
		needMultiplePages = false
		maxPages = 1
		effectiveLimit = 0 
		fmt.Printf("No post limit provided, fetching only first page for user %s\n", username)
		
	case limit == -1:
		needMultiplePages = true
		effectiveLimit = -1 
		if sinceTimestamp > 0 {
			maxPages = 1000 
			fmt.Printf("Fetching ALL posts for user %s since timestamp %d\n", username, sinceTimestamp)
		} else {
			maxPages = 500 
			fmt.Printf("Fetching ALL posts for user %s (no timestamp filter)\n", username)
		}
		
	default:
		needMultiplePages = limit > 25 || sinceTimestamp > 0
		maxPages = (limit/25 + 1) * 2 
		if maxPages > 50 {
			maxPages = 50 
		}
		effectiveLimit = limit
		fmt.Printf("Fetching up to %d posts for user %s\n", limit, username)
	}

	if sinceTimestamp > 0 {
		sinceTime := time.Unix(sinceTimestamp, 0)
		fmt.Printf("Filtering posts since %s (timestamp: %d)\n", sinceTime.Format(time.RFC3339), sinceTimestamp)
	}

	for pageCount < maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		pageCount++
		apiURL := s.client.GetUserPostsURL(username, after)
		fmt.Printf("Fetching posts page %d for user %s\n", pageCount, username)

		data, err := s.client.FetchJSON(ctx, apiURL)
		if err != nil {
			return nil, fmt.Errorf("fetch user posts: %w", err)
		}

		pagePosts, nextAfter, err := s.parser.ParseUserPosts(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("parse user posts: %w", err)
		}

		reachedTimeLimit := false
		pagePostCount := 0
		
		for _, post := range pagePosts {
		
			if sinceTimestamp > 0 && post.CreatedAt.Unix() < sinceTimestamp {
				reachedTimeLimit = true
				continue 
			}
			
			pagePostCount++
			posts = append(posts, post)
			

			if effectiveLimit > 0 && len(posts) >= effectiveLimit {
				fmt.Printf("Reached requested limit of %d posts\n", effectiveLimit)
				return posts, nil
			}
		}

		fmt.Printf("Posts page %d yielded %d posts (total now: %d)\n",
			pageCount, pagePostCount, len(posts))

		// Stop conditions
		if reachedTimeLimit && sinceTimestamp > 0 {
			fmt.Println("Reached timestamp cutoff, stopping pagination")
			break
		}

		if !needMultiplePages {
			fmt.Println("First page only mode, stopping pagination")
			break
		}

		if nextAfter == "" || pagePostCount == 0 {
			fmt.Println("No more posts available")
			break
		}

		after = nextAfter

		var timeoutDuration time.Duration
		if effectiveLimit == -1 {
			timeoutDuration = 5 * time.Minute 
		} else {
			timeoutDuration = 2 * time.Minute 
		}
		
		if time.Since(startTime) > timeoutDuration && len(posts) > 0 {
			fmt.Printf("Time limit (%v) reached, returning results so far\n", timeoutDuration)
			break
		}
		
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("Final result: %d posts fetched for user %s\n", len(posts), username)
	return posts, nil
}

//  fetchUserComments function
func (s *scraperService) fetchUserComments(
	ctx context.Context,
	username string,
	sinceTimestamp int64,
	limit int,
) ([]models.UserComment, error) {
	var comments []models.UserComment
	after := ""
	pageCount := 0
	startTime := time.Now()

	// Determine pagination behavior based on limit
	var needMultiplePages bool
	var maxPages int
	var effectiveLimit int

	switch {
	case limit == 0:
		needMultiplePages = false
		maxPages = 1
		effectiveLimit = 0 
		fmt.Printf("No comment limit provided, fetching only first page for user %s\n", username)
		
	case limit == -1:
		needMultiplePages = true
		effectiveLimit = -1 
		if sinceTimestamp > 0 {
			maxPages = 1000 
			fmt.Printf("Fetching ALL comments for user %s since timestamp %d\n", username, sinceTimestamp)
		} else {
			maxPages = 500 
			fmt.Printf("Fetching ALL comments for user %s (no timestamp filter)\n", username)
		}
		
	default:
		needMultiplePages = limit > 25 || sinceTimestamp > 0
		maxPages = (limit/25 + 1) * 2 // Estimate pages needed
		if maxPages > 50 {
			maxPages = 50 
		}
		effectiveLimit = limit
		fmt.Printf("Fetching up to %d comments for user %s\n", limit, username)
	}


	if sinceTimestamp > 0 {
		sinceTime := time.Unix(sinceTimestamp, 0)
		fmt.Printf("Filtering comments since %s (timestamp: %d)\n", sinceTime.Format(time.RFC3339), sinceTimestamp)
	}

	for pageCount < maxPages {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		pageCount++
		apiURL := s.client.GetUserCommentsURL(username, after)
		fmt.Printf("Fetching comments page %d for user %s\n", pageCount, username)

		data, err := s.client.FetchJSON(ctx, apiURL)
		if err != nil {
			return nil, fmt.Errorf("fetch user comments: %w", err)
		}

		pageComments, nextAfter, err := s.parser.ParseUserComments(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("parse user comments: %w", err)
		}

		reachedTimeLimit := false
		pageCommentCount := 0
		
		for _, comment := range pageComments {
			if sinceTimestamp > 0 && comment.CreatedAt.Unix() < sinceTimestamp {
				reachedTimeLimit = true
				continue 
			}
			
			pageCommentCount++
			comments = append(comments, comment)
			
			if effectiveLimit > 0 && len(comments) >= effectiveLimit {
				fmt.Printf("Reached requested limit of %d comments\n", effectiveLimit)
				return comments, nil
			}
		}

		fmt.Printf("Comments page %d yielded %d comments (total now: %d)\n",
			pageCount, pageCommentCount, len(comments))

		// Stop conditions
		if reachedTimeLimit && sinceTimestamp > 0 {
			fmt.Println("Reached timestamp cutoff, stopping pagination")
			break
		}

		if !needMultiplePages {
			fmt.Println("First page only mode, stopping pagination")
			break
		}

		if nextAfter == "" || pageCommentCount == 0 {
			fmt.Println("No more comments available")
			break
		}

		after = nextAfter
		var timeoutDuration time.Duration
		if effectiveLimit == -1 {
			timeoutDuration = 5 * time.Minute 
		} else {
			timeoutDuration = 2 * time.Minute 
		}
		
		if time.Since(startTime) > timeoutDuration && len(comments) > 0 {
			fmt.Printf("Time limit (%v) reached, returning results so far\n", timeoutDuration)
			break
		}
		
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("Final result: %d comments fetched for user %s\n", len(comments), username)
	return comments, nil
}

// ScrapePost retrieves a post with all its comments, including all "load more" content
func (s *scraperService) ScrapePost(ctx context.Context, postID string) (models.PostDetail, error) {
    startTime := time.Now()
    fmt.Printf("[%s] Starting to scrape post %s\n", startTime.Format(time.RFC3339), postID)

    // Fetch initial post with first level comments
    detail, err := s.fetchInitialPost(ctx, postID)
    if err != nil {
        return models.PostDetail{}, err
    }
    
    initialCommentCount := s.countComments(detail.Comments)
    fmt.Printf("Initial post fetch retrieved %d comments\n", initialCommentCount)


    // Expand all "load more" comment sections
    expandedCount := s.expandCommentsFast(ctx, postID, &detail)
    

    elapsed := time.Since(startTime)
    totalComments := s.countComments(detail.Comments)
    
    fmt.Printf("[%s] Finished scraping post %s in %v - found %d total comments (expanded %d)\n", 
        time.Now().Format(time.RFC3339), postID, elapsed, totalComments, expandedCount)
    
    return detail, nil
}

// fetchInitialPost retrieves the post with its initial comments
func (s *scraperService) fetchInitialPost(ctx context.Context, postID string) (models.PostDetail, error) {
    apiURL := s.client.GetPostURL(postID)
    data, err := s.client.FetchJSON(ctx, apiURL)
    if err != nil {
        return models.PostDetail{}, fmt.Errorf("fetch post JSON: %w", err)
    }

    var raw []json.RawMessage
    if err := json.Unmarshal(data, &raw); err != nil || len(raw) < 2 {
        return models.PostDetail{}, fmt.Errorf("invalid post JSON format: %w", err)
    }

    return s.parser.ParsePost(ctx, raw[0], raw[1])
}



// expandCommentsFast uses concurrent processing to load comments faster
func (s *scraperService) expandCommentsFast(ctx context.Context, postID string, detail *models.PostDetail) int {
    expandedCount := 0
    maxIterations := 60 
    
    workerCount := 3    
    
    remainingIDs := 0
    stuckCount := 0
    stuckLimit := 3      // Increased from 2
    
    for iteration := 0; iteration < maxIterations; iteration++ {
        moreSets := s.findMoreComments(ctx, detail)
        if len(moreSets) == 0 {
            fmt.Println("No more 'load more' comments found, expansion complete")
            break
        }
        
        newRemainingIDs := 0
        for _, set := range moreSets {
            newRemainingIDs += len(set.CommentIDs)
        }
        
        if newRemainingIDs == remainingIDs && newRemainingIDs > 0 {
            stuckCount++
            if stuckCount >= stuckLimit {
                fmt.Printf("No progress after %d iterations, stopping\n", stuckCount)
                break
            }
        } else {
            stuckCount = 0
        }
        remainingIDs = newRemainingIDs
        
        fmt.Printf("Iteration %d: Processing %d more comment sets (%d IDs remaining)\n", 
            iteration, len(moreSets), remainingIDs)
        
        // Add proper delay between iterations
        if iteration > 0 {
            time.Sleep(2 * time.Second)  // Increased delay
        }
        
        // Take longer breaks periodically
        if iteration > 10 && iteration % 5 == 0 {
            fmt.Println("Taking longer break after multiple iterations")
            time.Sleep(5 * time.Second)
        }
        
        batchSize := 15  // Reduced from 30
        if len(moreSets) > batchSize {
            fmt.Printf("Limiting to %d more comment sets per iteration\n", batchSize)
            moreSets = moreSets[:batchSize]
        }
        
        commentSets := make(chan struct{
            Set struct {
                Parent string
                CommentIDs []string
                Depth int
                PlaceholderID string
            }
            Index int
        }, len(moreSets))
        
        results := make(chan struct {
            Comments []models.Comment
            Set struct {
                Parent string
                CommentIDs []string
                Depth int
                PlaceholderID string
            }
            Index int
        }, len(moreSets))
        
        var wg sync.WaitGroup
        for w := 0; w < workerCount; w++ {
            wg.Add(1)
            go func(workerId int) {
                defer wg.Done()
                s.commentWorker(ctx, postID, commentSets, results)
            }(w)
        }
        
        for i, set := range moreSets {
            if len(set.CommentIDs) == 0 {
                continue
            }
            commentSets <- struct {
                Set struct {
                    Parent string
                    CommentIDs []string
                    Depth int
                    PlaceholderID string
                }
                Index int
            }{Set: set, Index: i}
        }
        close(commentSets)
        
        go func() {
            wg.Wait()
            close(results)
        }()
        
        iterationCount := 0
        processedResults := make([]struct {
            Comments []models.Comment
            Set struct {
                Parent string
                CommentIDs []string
                Depth int
                PlaceholderID string
            }
            Index int
        }, 0, len(moreSets))
        
        for result := range results {
            processedResults = append(processedResults, result)
        }
        
        sort.Slice(processedResults, func(i, j int) bool {
            return processedResults[i].Index < processedResults[j].Index
        })
        
        for _, result := range processedResults {
            if len(result.Comments) > 0 {
                iterationCount += len(result.Comments)
                s.placeComments(detail, result.Set, result.Comments)
            }
        }
        
        expandedCount += iterationCount
        fmt.Printf("Added %d comments (total: %d)\n", iterationCount, expandedCount)
        
        if iterationCount == 0 {
            fmt.Println("No new comments added in this iteration, may be stuck")
        }
    }
    
    s.cleanupMoreComments(detail)
    
    return expandedCount
}

// commentWorker processes comment sets in parallel
func (s *scraperService) commentWorker(
    ctx context.Context, 
    postID string,
    commentSets <-chan struct {
        Set struct {
            Parent string
            CommentIDs []string
            Depth int
            PlaceholderID string
        }
        Index int
    },
    results chan<- struct {
        Comments []models.Comment
        Set struct {
            Parent string
            CommentIDs []string
            Depth int
            PlaceholderID string
        }
        Index int
    },
) {
    for work := range commentSets {
        comments, _ := s.fetchMoreCommentsFast(ctx, postID, work.Set.CommentIDs)
        
        results <- struct {
            Comments []models.Comment
            Set struct {
                Parent string
                CommentIDs []string
                Depth int
                PlaceholderID string
            }
            Index int
        }{
            Comments: comments,
            Set: work.Set,
            Index: work.Index,
        }
    }
}

// fetchMoreCommentsFast is an optimized version with fewer retries and delays
func (s *scraperService) fetchMoreCommentsFast(ctx context.Context, postID string, commentIDs []string) ([]models.Comment, error) {
    // Smaller batch size - Reddit sometimes rejects large batches
    const batchSize = 100
    var allComments []models.Comment
    
    var validIDs []string
    for _, id := range commentIDs {
        if id != "continue" {
            // Ensure proper format without t1_ prefix
            validIDs = append(validIDs, strings.TrimPrefix(id, "t1_"))
        }
    }
    
    if len(validIDs) == 0 {
        return allComments, nil
    }
    
    // Add debugging
    if len(validIDs) > 0 {
        fmt.Printf("Fetching %d comment IDs (first few: %v)\n", 
            len(validIDs), validIDs[:min(3, len(validIDs))])
    }
    
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    maxConcurrent := 2
    semaphore := make(chan struct{}, maxConcurrent)
    
    for i := 0; i < len(validIDs); i += batchSize {
        // Add delay between batches to avoid rate limiting
        if i > 0 {
            time.Sleep(1000 * time.Millisecond)
        }
        
        end := min(i+batchSize, len(validIDs))
        batch := validIDs[i:end]
        
        wg.Add(1)
        
        go func(batch []string, batchNum int) {
            defer wg.Done()
            
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            processedIDs := s.processCommentIDs(batch, len(batch))
            if len(processedIDs) == 0 {
                return
            }
            
            data, err := s.client.FetchMoreComments(ctx, postID, processedIDs)
            if err != nil {
                fmt.Printf("Error fetching comments batch %d: %v\n", batchNum, err)
                return
            }
            
            comments, err := s.parser.ParseMoreComments(ctx, data)
            if err != nil {
                fmt.Printf("Error parsing comments batch %d: %v\n", batchNum, err)
                return
            }
            
            if len(comments) > 0 {
                mu.Lock()
                allComments = append(allComments, comments...)
                mu.Unlock()
                
                fmt.Printf("Batch %d: retrieved %d comments\n", batchNum, len(comments))
            } else {
                fmt.Printf("Batch %d: WARNING - retrieved 0 comments for %d IDs\n", 
                    batchNum, len(processedIDs))
            }
        }(batch, i/batchSize)
    }
    
    wg.Wait()
    
    // Log results
    if len(allComments) == 0 && len(validIDs) > 0 {
        fmt.Printf("WARNING: No comments returned for %d IDs\n", len(validIDs))
    }
    
    return allComments, nil
}

func (s *scraperService) findMoreComments(ctx context.Context, detail *models.PostDetail) []struct {
    Parent string
    CommentIDs []string
    Depth int
    PlaceholderID string
} {
    var result []struct {
        Parent string
        CommentIDs []string
        Depth int
        PlaceholderID string
    }
    
    visited := make(map[string]bool)
    
    var traverse func(comments []models.Comment, parentID string, depth int)
    traverse = func(comments []models.Comment, parentID string, depth int) {
        for i := range comments {
            if ctx.Err() != nil {
                return
            }
            
            if comments[i].IsMore && len(comments[i].MoreIDs) > 0 && !visited[comments[i].ID] {
                visited[comments[i].ID] = true
                
                hasContinue := false
                for _, id := range comments[i].MoreIDs {
                    if id == "continue" {
                        hasContinue = true
                        fmt.Println("Found a 'continue' link, special handling might be needed")
                        break
                    }
                }
                
                if !hasContinue {
                    result = append(result, struct {
                        Parent string
                        CommentIDs []string
                        Depth int
                        PlaceholderID string
                    }{
                        Parent: parentID,
                        CommentIDs: comments[i].MoreIDs,
                        Depth: depth,
                        PlaceholderID: comments[i].ID,
                    })
                }
            }
            
            if comments[i].HasMore && len(comments[i].MoreIDs) > 0 && !visited[comments[i].ID+"-more"] {
                visited[comments[i].ID+"-more"] = true
                
                result = append(result, struct {
                    Parent string
                    CommentIDs []string
                    Depth int
                    PlaceholderID string
                }{
                    Parent: comments[i].ID,
                    CommentIDs: comments[i].MoreIDs,
                    Depth: depth + 1,
                    PlaceholderID: comments[i].ID + "-more",
                })
            }
            
            if len(comments[i].Replies) > 0 {
                traverse(comments[i].Replies, comments[i].ID, depth+1)
            }
        }
    }
    
    traverse(detail.Comments, detail.Post.ID, 0)
    
    sort.Slice(result, func(i, j int) bool {
        return len(result[i].CommentIDs) > len(result[j].CommentIDs)
    })
    
    return result
}

func (s *scraperService) processCommentIDs(ids []string, limit int) []string {
    if len(ids) == 0 {
        return nil
    }
    
    // Create a set to avoid duplicate IDs
    uniqueIDs := make(map[string]bool)
    
    result := make([]string, 0, min(len(ids), limit))
    for _, id := range ids {
        if len(result) >= limit {
            break
        }
        
        // Skip continue links
        if id == "continue" {
            continue
        }
        
        // Clean up ID format - ensure no t1_ prefix
        cleanID := strings.TrimPrefix(id, "t1_")
        
        // Check for duplicates
        if !uniqueIDs[cleanID] {
            uniqueIDs[cleanID] = true
            result = append(result, cleanID)
        }
    }
    
    return result
}
// placeComments - modified to sort comments and deduplicate more safely
func (s *scraperService) placeComments(detail *models.PostDetail, set struct {
    Parent string
    CommentIDs []string
    Depth int
    PlaceholderID string
}, newComments []models.Comment) {
    if len(newComments) == 0 {
        return
    }
    
    idMap := make(map[string]bool)
    var uniqueBatchComments []models.Comment
    
    for _, comment := range newComments {
        if !idMap[comment.ID] {
            idMap[comment.ID] = true
            uniqueBatchComments = append(uniqueBatchComments, comment)
        } else {
            fmt.Printf("Skipping duplicate within batch, ID: %s\n", comment.ID)
        }
    }
    
    // Sort by creation time
    sort.Slice(uniqueBatchComments, func(i, j int) bool {
        return uniqueBatchComments[i].CreatedAt.After(uniqueBatchComments[j].CreatedAt)
    })
    
    var placed bool
    
    if set.Depth == 0 {
        // Handle top-level comments
        placed = s.replacePlaceholder(&detail.Comments, set.PlaceholderID, uniqueBatchComments)
        if !placed {
            fmt.Printf("No placeholder found, adding %d comments to top level\n", len(uniqueBatchComments))
            
            // Deduplicate before adding to top level
            s.addWithoutDuplicates(&detail.Comments, uniqueBatchComments)
        }
    } else {
        // Handle nested comments
        placed = s.replaceInTree(&detail.Comments, set.Parent, set.PlaceholderID, uniqueBatchComments)
        
        // If failed, try to append to parent
        if !placed {
            placed = s.appendToParent(&detail.Comments, set.Parent, uniqueBatchComments)
            
            // If still failed, add to top level as last resort
            if !placed {
                fmt.Printf("WARNING: Could not find parent %s, adding %d comments to top level\n", 
                    set.Parent, len(uniqueBatchComments))
                
                // Deduplicate before adding to top level
                s.addWithoutDuplicates(&detail.Comments, uniqueBatchComments)
            }
        }
    }
}


func (s *scraperService) addWithoutDuplicates(existingComments *[]models.Comment, newComments []models.Comment) {
    // Create lookup map of existing comment IDs
    existingIDs := make(map[string]bool)
    for _, comment := range *existingComments {
        existingIDs[comment.ID] = true
    }
    
    // Only add comments that don't already exist
    addedCount := 0
    for _, comment := range newComments {
        if !existingIDs[comment.ID] {
            *existingComments = append(*existingComments, comment)
            addedCount++
        } else {
            fmt.Printf("Skipping already existing comment ID: %s\n", comment.ID)
        }
    }  
}

//appendToParent function that prevents duplicates
func (s *scraperService) appendToParent(comments *[]models.Comment, parentID string, newComments []models.Comment) bool {
    for i := range *comments {
        if (*comments)[i].ID == parentID {
            s.addWithoutDuplicates(&(*comments)[i].Replies, newComments)
            return true
        }
        
        if len((*comments)[i].Replies) > 0 {
            if s.appendToParent(&(*comments)[i].Replies, parentID, newComments) {
                return true
            }
        }
    }
    return false
}

// replacePlaceholder replaces a placeholder comment with actual comments
func (s *scraperService) replacePlaceholder(comments *[]models.Comment, placeholderID string, newComments []models.Comment) bool {
    if len(*comments) == 0 || len(newComments) == 0 {
        return false
    }
    
    for i := range *comments {
        if (*comments)[i].ID == placeholderID && (*comments)[i].IsMore {
            if i == len(*comments)-1 {
                *comments = append((*comments)[:i], newComments...)
            } else {
                *comments = append(append((*comments)[:i], newComments...), (*comments)[i+1:]...)
            }
            return true
        }
    }
    return false
}

// replaceInTree replaces a placeholder in a comment tree
func (s *scraperService) replaceInTree(comments *[]models.Comment, parentID, placeholderID string, newComments []models.Comment) bool {
    if len(*comments) == 0 || len(newComments) == 0 {
        return false
    }
    
    for i := range *comments {
        if (*comments)[i].ID == parentID {
            if s.replacePlaceholder(&(*comments)[i].Replies, placeholderID, newComments) {
                return true
            }
        }
        
        if len((*comments)[i].Replies) > 0 {
            if s.replaceInTree(&(*comments)[i].Replies, parentID, placeholderID, newComments) {
                return true
            }
        }
    }
    return false
}

// cleanupMoreComments removes remaining "more" placeholders
func (s *scraperService) cleanupMoreComments(detail *models.PostDetail) int {
    if len(detail.Comments) == 0 {
        return 0
    }
    
    removed := s.filterComments(&detail.Comments)
    
    for i := range detail.Comments {
        if len(detail.Comments[i].Replies) > 0 {
            removed += s.filterComments(&detail.Comments[i].Replies)
        }
    }
    
    return removed
}

// filterComments removes "more" placeholders from a comment list
func (s *scraperService) filterComments(comments *[]models.Comment) int {
    if len(*comments) == 0 {
        return 0
    }
    
    var filtered []models.Comment
    removed := 0
    
    for _, comment := range *comments {
        if !comment.IsMore {
            if len(comment.Replies) > 0 {
                removed += s.filterComments(&comment.Replies)
            }
            filtered = append(filtered, comment)
        } else {
            removed++
        }
    }
    
    *comments = filtered
    return removed
}

// countComments counts the total number of comments in a tree
func (s *scraperService) countComments(comments []models.Comment) int {
    if len(comments) == 0 {
        return 0
    }
    
    count := len(comments)
    
    for i := range comments {
        if len(comments[i].Replies) > 0 {
            count += s.countComments(comments[i].Replies)
        }
    }
    
    return count
}

// Utility function for min
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// Search function for Reddit content
func (s *scraperService) Search(
	ctx context.Context,
	searchParams map[string]string,
	sinceTimestamp int64,
	limit int,
) ([]models.Post, error) {
	startTime := time.Now()
	var posts []models.Post

	if limit == -1 && sinceTimestamp == 0 {
		limit = 1000 
		fmt.Printf("Limit was -1 with no timestamp filter for search, using default limit of %d\n", limit)
	}

	apiLimit := 100 
	
	
	if limit > 0 && limit < apiLimit {
		apiLimit = limit
	}


	searchParams["limit"] = strconv.Itoa(apiLimit)

	after := ""
	pageCount := 0
	maxPages := 10 

	if limit == -1 && sinceTimestamp > 0 {
		maxPages = 1000 
		fmt.Printf("Fetching all search results until timestamp %d\n", sinceTimestamp)
	} else if limit > 0 {
		// Estimate pages needed based on limit
		estimatedPages := (limit + apiLimit - 1) / apiLimit 
		maxPages = estimatedPages * 2 
		fmt.Printf("Fetching up to %d search results (estimated %d pages)\n", limit, estimatedPages)
	}

	for pageCount < maxPages {
		if ctx.Err() != nil {
			return posts, ctx.Err()
		}

		pageCount++


		if after != "" {
			searchParams["after"] = after
		} else {
			delete(searchParams, "after")
		}

		apiURL := s.client.GetSearchURL(searchParams)
		fmt.Printf("Fetching search page %d\n", pageCount)

		data, err := s.client.FetchJSON(ctx, apiURL)
		if err != nil {
			return nil, fmt.Errorf("fetch search results: %w", err)
		}

		pagePosts, nextAfter, err := s.parser.ParseSubreddit(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("parse search results: %w", err)
		}

		pagePostCount := 0
		reachedTimeLimit := false


		for _, post := range pagePosts {
			if sinceTimestamp > 0 && post.CreatedAt.Unix() < sinceTimestamp {
				reachedTimeLimit = true
				continue
			}

			pagePostCount++
			posts = append(posts, post)
		}

		fmt.Printf("Search page %d yielded %d posts (total now: %d/%d)\n",
			pageCount, pagePostCount, len(posts), limit)

		if limit > 0 && len(posts) >= limit {
			fmt.Println("Reached requested limit, stopping pagination")
			break
		}

		if reachedTimeLimit && sinceTimestamp > 0 {
			fmt.Println("Reached time limit cutoff, stopping pagination")
			break
		}

		if nextAfter == "" || pagePostCount == 0 {
			fmt.Println("No more pages available or empty page")
			break
		}

		after = nextAfter

		timeoutDuration := 60 * time.Second
		if limit == -1 {
			timeoutDuration = 3 * time.Minute
		}
		
		if time.Since(startTime) > timeoutDuration && len(posts) > 0 {
			fmt.Printf("Time limit (%v) reached, returning results so far\n", timeoutDuration)
			break
		}
		

		time.Sleep(200 * time.Millisecond)
	}

	if limit > 0 && len(posts) > limit {
		posts = posts[:limit]
	}

	fmt.Printf("Final search result: %d posts fetched in %v\n", len(posts), time.Since(startTime))
	return posts, nil
}