# Usage — `reddit-ingestion`

This document provides detailed instructions and examples for using the `reddit-ingestion` API. It includes endpoint specifications, parameter descriptions, and common usage patterns.

> **Note**: Interactive API documentation is available via Swagger UI at `/swagger/index.html` when the server is running. This provides a convenient way to explore and test all endpoints.
---

## API Endpoints Overview

| Endpoint       | Purpose                                        | Key Parameters                         |
|----------------|------------------------------------------------|----------------------------------------|
| `/subreddit`   | Fetch posts from a specific subreddit          | `subreddit`, `limit`, `since_timestamp` |
| `/user`        | Get user information, posts, and comments      | `username`, `post_limit`, `comment_limit` |
| `/post`        | Get a post with all its comments               | `post_id`                               |
| `/search`      | Search Reddit content with filters             | `search_string`, `subreddit`, `author`   |
| `/health`      | Check service health                           | None                                    |

---

## Endpoint: `/subreddit`

Retrieves posts from a specified subreddit with pagination support.

### Parameters

| Parameter         | Required | Description                                      | Default |
|-------------------|----------|--------------------------------------------------|---------|
| `subreddit`       | Yes      | Subreddit name (without "r/")                    | None    |
| `limit`           | No       | Maximum number of posts to retrieve              | 25      |
| `since_timestamp` | No       | Only return posts newer than this Unix timestamp | 0       |

### Special Values

- `limit=-1`: Retrieve all posts (use with caution)
- `limit=0`: Use default limit (25)

### Example

```
GET /subreddit?subreddit=golang&limit=10
```

### Response

```json
{
  "posts": [
    {
      "id": "abcd123",
      "title": "Go 1.22 Released",
      "body": "Text content of the post...",
      "author": "gopher",
      "score": 42,
      "created_at": "2025-04-15T12:00:00Z",
      "flair": "News",
      "url": "https://reddit.com/r/golang/comments/abcd123/go_119_released/"
    },
    ...
  ],
  "meta": {
    "requested_limit": 10,
    "actual_count": 10,
    "subreddit": "golang",
    "since_timestamp": 0,
    "processing_time_ms": 1250
  }
}
```

---

## Endpoint: `/user`

Retrieves information about a Reddit user, including profile details, posts, and comments.

### Parameters

| Parameter         | Required | Description                                      | Default |
|-------------------|----------|--------------------------------------------------|---------|
| `username`        | Yes      | Reddit username                                  | None    |
| `post_limit`      | No       | Maximum number of posts to retrieve              | 25      |
| `comment_limit`   | No       | Maximum number of comments to retrieve           | 25      |
| `since_timestamp` | No       | Only return content newer than this timestamp    | 0       |

### Special Values

- `post_limit=-1` or `comment_limit=-1`: Retrieve all posts/comments
- `post_limit=0` or `comment_limit=0`: Use default limits

### Example

```
GET /user?username=spez&post_limit=5&comment_limit=10
```

### Response

```json
{
  "user_info": {
    "username": "spez",
    "link_karma": 15983,
    "comment_karma": 28450,
    "created_at": "2005-06-06T04:01:40Z"
  },
  "posts": [
    {
      "id": "xyz789",
      "title": "Introducing Reddit Talk",
      "body": "Today we're rolling out Reddit Talk...",
      "score": 9876,
      "created_at": "2025-04-10T16:30:00Z",
      "subreddit": "blog",
      "url": "https://reddit.com/r/blog/comments/xyz789/introducing_reddit_talk/",
      "flair": "Announcement"
    },
    ...
  ],
  "comments": [
    {
      "id": "def456",
      "body": "We're working on fixing that issue...",
      "score": 532,
      "created_at": "2025-04-12T14:25:10Z",
      "subreddit": "announcements",
      "post_id": "uvw345",
      "post_title": "An update on Reddit's policies"
    },
    ...
  ]
}
```

---

## Endpoint: `/post`

Retrieves a post with all its comments, including "load more" content.

### Parameters

| Parameter  | Required | Description                | Default |
|------------|----------|----------------------------|---------|
| `post_id`  | Yes      | Reddit post ID (not URL)   | None    |

### Example

```
GET /post?post_id=abc123
```

### Response

```json
{
  "post": {
    "id": "abc123",
    "title": "What's your favorite Go framework?",
    "body": "I'm starting a new project and wondering what framework to use...",
    "author": "coder123",
    "score": 25,
    "created_at": "2025-04-14T09:15:00Z",
    "flair": "Question",
    "url": "https://reddit.com/r/golang/comments/abc123/whats_your_favorite_go_framework/"
  },
  "comments": [
    {
      "id": "comment1",
      "author": "dev456",
      "body": "I prefer Echo for its simplicity",
      "score": 18,
      "created_at": "2025-04-14T09:30:00Z",
      "replies": [
        {
          "id": "reply1",
          "author": "webdev789",
          "body": "Echo is great! I use it for all my projects.",
          "score": 7,
          "created_at": "2025-04-14T10:05:00Z",
          "replies": []
        }
      ]
    },
    ...
  ]
}
```

---

## Endpoint: `/search`

Searches Reddit content with various filters.

### Parameters

| Parameter         | Required | Description                                      | Default     |
|-------------------|----------|--------------------------------------------------|-------------|
| `search_string`   | Yes      | Text to search for                               | None        |
| `subreddit`       | No       | Limit search to specific subreddit               | None        |
| `author`          | No       | Limit search to specific author                  | None        |
| `sort`            | No       | Sort order (`relevance`, `new`, `top`, etc.)     | `relevance` |
| `time`            | No       | Time range (`hour`, `day`, `week`, `month`, `year`, `all`) | `all` |
| `limit`           | No       | Maximum number of results                        | 25          |
| `since_timestamp` | No       | Only return content newer than this timestamp    | 0           |

### Example

```
GET /search?search_string=golang+tutorial&subreddit=golang&sort=new&limit=10
```

### Alternative Compound Query

You can also use the compound query format:

```
GET /search?compound_query=golang+tutorial+subreddit:golang&sort=new&limit=10
```

### Response

```json
{
  "posts": [
    {
      "id": "abc456",
      "title": "Comprehensive Go Tutorial for Beginners",
      "body": "I've created a new tutorial series...",
      "author": "goteacher",
      "score": 156,
      "created_at": "2025-04-13T18:20:00Z",
      "flair": "Tutorial",
      "url": "https://reddit.com/r/golang/comments/abc456/comprehensive_go_tutorial_for_beginners/"
    },
    ...
  ],
  "meta": {
    "query": "golang tutorial",
    "params": {
      "search_string": "golang tutorial",
      "subreddit": "golang",
      "sort": "new"
    },
    "count": 10,
    "processing_time_ms": 1800,
    "requested_limit": 10
  }
}
```

---

## Common Usage Patterns

### 1. Monitoring a Subreddit for New Posts

To check for new posts since your last check:

```
GET /subreddit?subreddit=golang&since_timestamp=1675423800
```

### 2. Full User Activity Analysis

To get all of a user's posts and comments:

```
GET /user?username=gopher123&post_limit=-1&comment_limit=-1
```

### 3. Real-time Post Tracking

To get the latest state of a post with all comments:

```
GET /post?post_id=abc123
```

### 4. Keyword Monitoring Across Multiple Subreddits

Using search to find mentions of specific keywords:

```
GET /search?search_string=go+programming+language&sort=new&limit=50
```

---

## Rate Limiting Considerations

- The service uses proxies to avoid Reddit's rate limits, but has its own limits
- For production use, consider:
  - Implementing client-side throttling
  - Caching frequently accessed data
  - Using reasonable limits rather than `-1` for unlimited fetching
  - Spacing out requests for large data sets

---

## Error Responses

| Status Code | Description                 | Example Cause                          |
|-------------|-----------------------------|----------------------------------------|
| 400         | Bad Request                 | Missing required parameter             |
| 404         | Not Found                   | Subreddit or user doesn't exist        |
| 429         | Too Many Requests           | Rate limited by Reddit                 |
| 502         | Bad Gateway                 | Error communicating with Reddit API    |
| 504         | Gateway Timeout             | Reddit API took too long to respond    |

---

## Related Docs

- [Getting Started](./getting-started.md)
- [Troubleshooting](./troubleshooting.md)