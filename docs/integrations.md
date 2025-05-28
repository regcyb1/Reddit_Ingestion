# Integrations — `reddit-ingestion`

This document outlines the external dependencies and integration points of the `reddit-ingestion` service. These integrations are essential for data collection, request handling, and system operation.

---

## External Dependencies

### 1. Reddit API (Old Reddit)

**Purpose**: Primary data source for all content ingestion.

**Integration Details**:
- Base URL: `https://old.reddit.com` (configurable via `REDDIT_BASE_URL`)
- API Endpoints Used:
  - Subreddit posts: `/r/{subreddit}/new.json`
  - User data: `/user/{username}/about.json`
  - User posts: `/user/{username}/submitted/new.json`
  - User comments: `/user/{username}/comments/.json`
  - Post data: `/comments/{postID}.json`
  - Search: `/search.json`
  - More comments: `https://api.reddit.com/api/morechildren`

**Authentication**: None (public API access)

---

### 2. HTTP/HTTPS Proxies

**Purpose**: Intermediary for Reddit API requests to avoid rate limiting and IP bans.

**Integration Details**:
- Multiple proxies supported via `REDDIT_PROXY_URLS`
- Automatic rotation between requests
- Supports HTTP and HTTPS proxies with authentication
- TLS fingerprinting to mimic legitimate browsers

**Connection Method**: Direct HTTP/HTTPS requests through proxy servers


---

## Architecture Diagram

```
┌─────────────────┐      ┌─────────────┐      ┌──────────────┐
│                 │      │             │      │              │
│  HTTP Clients   ├─────►│  API Server ├─────►│  Scraper     │
│  (Users)        │      │  (Echo)     │      │  Service     │
│                 │      │             │      │              │
└─────────────────┘      └─────────────┘      └──────┬───────┘
                                                     │
                                                     ▼
                                            ┌────────────────┐
                                            │                │
                                            │  Proxy Rotator │
                                            │                │
                                            └──────┬─────────┘
                                                   │
                                     ┌─────────────┴─────────────┐
                                     │                           │
                                  ┌──┴───┐                    ┌──┴───┐
                                  │Proxy1│  ...  Proxies  ... │ProxyN│
                                  └──────┘                    └──────┘
                                     │                           │
                                     └─────────────┬─────────────┘
                                                   │
                                                   ▼
                                             ┌──────────┐
                                             │          │
                                             │  Reddit  │
                                             │          │
                                             └──────────┘
```

---

## Integration Considerations

| Integration      | Type      | Required | Notes                                |
|------------------|-----------|----------|--------------------------------------|
| Reddit API       | External  | Yes      | Primary data source                  |
| HTTP Proxies     | External  | Yes      | For avoiding rate limits             |
| Prometheus       | Optional  | No       | For metrics collection               |
| Docker           | Optional  | No       | For containerized deployment         |


---

## Reddit API Limitations

- Reddit enforces rate limits per IP address
- Old Reddit API is more reliable for scraping than the new Reddit API
- Some endpoints may be temporarily unavailable during Reddit maintenance
- Complex comment trees require multiple API calls to fully expand

---

## Proxy Requirements

For production use, proxies should:
- Support HTTP/HTTPS
- Have stable uptime
- Rotate IPs frequently or have a large pool
---

## Related Docs

- [Configuration](./configuration.md)
- [Deployment](./deployment.md)
- [Observability](./observability.md)