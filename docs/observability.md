# Observability — `reddit-ingestion`

This document explains how to monitor Reddit ingestion service.

---

## Logging

Used structured JSON logging via Echo's built-in logger middleware:

- API requests and responses
- Proxy selection events
- Reddit API errors
- Retries and backoff attempts
- Parsing errors

Example log output:

```json
{
  "time": "2025-04-17T15:04:05Z",
  "level": "info",
  "prefix": "reddit-ingestion",
  "file": "scraper/service.go",
  "line": "256",
  "message": "Fetching page 1 for subreddit golang"
}
```

---

## Health Check

Implemented a `/health` endpoint that returns HTTP 200 when the service is healthy. This endpoint checks:

- Server is running 
- Basic proxy connectivity
- Reddit API connectivity

This endpoint is compatible with Docker healthchecks.

---

## What to Monitor

Recommend monitoring the following:

1. **API Request Rate**: Check requests per second to each endpoint
2. **Error Rate**: Monitor percentage of requests returning errors
3. **Response Time**: Watch for latency increases
4. **Proxy Status**: Monitor for proxy connection issues
5. **Memory Usage**: Track container memory consumption

---

## Troubleshooting with Logs

When troubleshooting, I search for these log patterns:

| Issue                           | Log Pattern to Search                              |
|---------------------------------|---------------------------------------------------|
| Reddit API Rate Limit           | `rate limit exceeded` or `status code 429`         |
| Proxy Connection Failure        | `failed to connect via proxy`                      |
| Parser Errors                   | `parse error` or `unmarshaling JSON`               |
| Memory Issues                   | `runtime: out of memory`                           |

---
