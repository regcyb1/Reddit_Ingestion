# Service Requirement Document — `reddit-ingestion`

---

## Service Overview

The `reddit-ingestion` service is responsible for reliably fetching and structuring data from Reddit. It handles various content types including subreddit posts, user activities, post details with comments, and search results. The service ensures consistent data collection through proxy rotation, browser fingerprinting, and resilient retries, effectively managing Reddit's rate limiting policies.

---

## Purpose

To provide a reliable, scalable interface for extracting Reddit content with configurable limits and filters while avoiding rate limiting and blocks through proxy rotation and browser fingerprinting techniques.

---

## API Endpoints

| Endpoint       | HTTP Method | Parameters                              | Description                           |
|----------------|-------------|------------------------------------------|---------------------------------------|
| `/subreddit`   | GET         | `subreddit`, `limit`, `since_timestamp` | Fetch posts from a subreddit          |
| `/user`        | GET         | `username`, `post_limit`, `comment_limit`, `since_timestamp` | Get user activity |
| `/post`        | GET         | `post_id`                               | Get post details with all comments    |
| `/search`      | GET         | `search_string`, `subreddit`, `author`, etc. | Search Reddit with filters        |
| `/health`      | GET         | None                                    | Service health check                  |

---

## Data Models

### Post Data

```json
{
  "id": "abcd123",
  "title": "Post title",
  "body": "Post content",
  "author": "username",
  "score": 42,
  "created_at": "2025-04-17T15:04:05Z",
  "flair": "Discussion",
  "url": "https://reddit.com/r/example/comments/abcd123/post_title/"
}
```

### Comment Data

```json
{
  "id": "efgh456",
  "author": "username",
  "body": "Comment content",
  "score": 10,
  "created_at": "2025-04-17T15:14:25Z",
  "replies": [
    {
      "id": "ijkl789",
      "author": "another_user",
      "body": "Reply content",
      "score": 5,
      "created_at": "2025-04-17T15:20:05Z",
      "replies": []
    }
  ]
}
```

### User Data

```json
{
  "user_info": {
    "username": "example_user",
    "link_karma": 1000,
    "comment_karma": 5000,
    "created_at": "2020-01-01T00:00:00Z"
  },
  "posts": [...],
  "comments": [...]
}
```

---

## Technical Requirements

### Proxy Management

- Support for multiple HTTP/HTTPS proxies
- Automatic rotation between proxies
- Handling of proxy authentication
- Masking of proxy credentials in logs

### Browser Fingerprinting

- Simulation of common browser fingerprints (Chrome, Firefox, Safari, Edge)
- Customizable User-Agent strings
- TLS fingerprinting to avoid detection

### Resilient Fetching

- Configurable retry mechanism with exponential backoff
- Handling of Reddit-specific error responses
- Support for request timeouts
- Rate limiting awareness

### Performance

- Support for concurrent scraping operations
- Efficient handling of nested comment structures
- Memory-efficient processing of large responses
- Configurable request limits and timeouts

---

## System Architecture

```
                                 ┌───────────────────────┐
                                 │                       │
                                 │     API Layer         │
                                 │    (Echo Server)      │
                                 │                       │
                                 └───────────┬───────────┘
                                             │
                 ┌───────────────────────────┴────────────────────────┐
                 │                                                    │
    ┌────────────▼─────────────┐            ┌─────────────────────────▼─┐
    │                          │            │                           │
    │    Subreddit Handler     │            │      User Handler         │
    │                          │            │                           │
    └────────────┬─────────────┘            └─────────────┬─────────────┘
                 │                                        │
                 │              ┌─────────────────────────┴─┐
                 │              │                           │
                 │              │      Post Handler         │
                 │              │                           │
                 │              └─────────────┬─────────────┘
                 │                            │
                 │              ┌─────────────┴─────────────┐
                 │              │                           │
                 │              │     Search Handler        │
                 │              │                           │
                 │              └─────────────┬─────────────┘
                 │                            │
┌────────────────┴────────────────────────────┴─────────────────────────┐
│                                                                       │
│                        Scraper Service                                │
│                                                                       │
└────────────────────────────────┬─────────────────────────────────────-┘
                                 │
                 ┌───────────────┴───────────────┐
                 │                               │
    ┌────────────▼─────────────┐   ┌─────────────▼─────────────┐
    │                          │   │                           │
    │     Reddit Client        │   │        Parser             │
    │                          │   │                           │
    └────────────┬─────────────┘   └───────────────────────────┘
                 │
                 │
    ┌────────────▼─────────────┐
    │                          │
    │     Proxy Rotator        │
    │                          │
    └──────────────────────────┘
```

---

## Error Handling

| Error Type                      | Handling Strategy                                |
|---------------------------------|--------------------------------------------------|
| Reddit API Rate Limiting (429)  | Retry with different proxy after backoff         |
| Reddit API Unavailable (503)    | Exponential backoff and retry                    |
| Proxy Connection Failure        | Switch to alternate proxy immediately            |
| Parsing Errors                  | Log error details and return partial data if possible |
| Timeout Errors                  | Cancel context and retry with longer timeout     |
| Invalid Parameters              | Return clear HTTP 400 error with explanation     |

---

## Security Considerations

- Proxy credentials are masked in logs
- No write operations to Reddit
- TLS is used for all external connections
- Secrets management for proxy credentials
- No user authentication data is collected

---

## Scaling Considerations

- Service is stateless and can be horizontally scaled
- Use more proxies as request volume increases
- Configure timeouts appropriate to instance count
- Consider using a caching layer for popular requests
- Memory usage scales with request complexity (e.g., number of comments)

---

## Monitoring Requirements

- API request/response metrics
- Proxy usage and rotation metrics 
- Reddit API error counts
- Memory and CPU usage
- Request latency by endpoint

---

## References

- [Reddit API Documentation](https://www.reddit.com/dev/api/)
- [Rate Limiting Policies](https://github.com/reddit-archive/reddit/wiki/API)