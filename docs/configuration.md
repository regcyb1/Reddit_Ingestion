# Configuration — `reddit-ingestion`

This document lists all environment variables to use in Reddit ingestion service.

---

## Required Variables

| Variable            | Description                                           | Example                           |
|---------------------|-------------------------------------------------------|-----------------------------------|
| `REDDIT_PROXY_URLS` | Comma-separated list of HTTP/HTTPS proxy URLs         | `http://user:pass@proxy.com:8080,http://user:pass@proxy2.com:8080` |
| `REDDIT_USER_AGENT` | User agent string for Reddit API requests             | `Mozilla/5.0`    |

---

## Optional Variables

| Variable                   | Description                                      | Default       | Example              |
|----------------------------|--------------------------------------------------|---------------|----------------------|
| `PROXY_MAX_RETRIES`        | Number of retry attempts for failed requests     | `3`           | `5`                  |
| `SERVER_PORT`              | Port for the API server                          | `8080`        | `9000`               |
| `REDDIT_BASE_URL`          | Base URL for Reddit API                          | `https://old.reddit.com` | `https://reddit.com` |
| `SCRAPER_DEFAULT_POST_LIMIT` | Default limit for post fetching                | `25`          | `50`                 |
| `SCRAPER_DEFAULT_COMMENT_LIMIT` | Default limit for comment fetching          | `50`          | `100`                |

---

## Proxy Configuration

You need at least one proxy URL. recommend using multiple proxies for better reliability:

```
REDDIT_PROXY_URLS=http://user:pass@proxy1.com:8080,http://user:pass@proxy2.com:8080
```

Mask proxy credentials in logs for security.

---

## Example Configuration

Here's what `.env` file typically looks like:

```text
# Server configuration
SERVER_PORT=8080

# Reddit API configuration
REDDIT_USER_AGENT=Mozilla/5.0
REDDIT_BASE_URL=https://old.reddit.com

# Proxy configuration
REDDIT_PROXY_URLS=http://user:pass@proxy1.com:12233
PROXY_MAX_RETRIES=3

# Scraper configuration
SCRAPER_DEFAULT_POST_LIMIT=25 
SCRAPER_DEFAULT_COMMENT_LIMIT=50
```