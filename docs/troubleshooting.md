# Troubleshooting — `reddit-ingestion`

This document outlines common problems encountered with the `reddit-ingestion` service and provides diagnostic steps and solutions.

---

## Table of Contents

- [Proxy Connection Issues](#proxy-connection-issues)
- [Rate Limiting Problems](#rate-limiting-problems)
- [Comment Retrieval Failures](#comment-retrieval-failures)
- [Memory Consumption Issues](#memory-consumption-issues)
- [Slow Response Times](#slow-response-times)
- [Parser Errors](#parser-errors)
- [API Usage Problems](#api-usage-problems)
- [Testing Issues](#testing-issues)

---

## Proxy Connection Issues

**Symptoms**:
- Logs show `dial tcp: i/o timeout` or `no route to host` errors
- All requests fail with connection errors
- `/health` endpoint is returning errors

**Possible Causes**:
- Invalid proxy URLs
- Proxies are offline or unreachable
- Proxy authentication failure
- Network connectivity issues

**Diagnostic Steps**:
1. Verify proxy URLs are correctly formatted:
   ```
   http://username:password@host:port
   ```

2. Test proxies manually:
   ```bash
   curl -x http://username:password@host:port https://httpbin.org/ip
   ```

3. Check logs for specific proxy errors:
   ```bash
   grep "proxy" /var/log/reddit-ingestion.log
   ```

**Solutions**:
- Update proxy URLs in configuration
- Replace unreliable proxies
- Ensure network allows outbound connections to proxy servers
- Verify proxy authentication credentials

---

## Rate Limiting Problems

**Symptoms**:
- Logs show HTTP 429 (Too Many Requests) responses
- Excessive proxy rotations
- Scraping operations incomplete or slow

**Possible Causes**:
- Too many requests to Reddit in a short time
- Not enough proxies in the rotation
- Proxies being detected by Reddit
- `RATE_LIMIT_DELAY` set too low

**Diagnostic Steps**:
1. Check logs for rate limit responses:
   ```bash
   grep "429" /var/log/reddit-ingestion.log
   ```

2. Monitor proxy rotation frequency:
   ```bash
   grep "Switching to proxy" /var/log/reddit-ingestion.log | wc -l
   ```

3. Verify user agent diversity.

**Solutions**:
- Increase `RATE_LIMIT_DELAY` to slow down requests
- Add more proxies to `REDDIT_PROXY_URLS`
- Use residential proxies instead of datacenter proxies
- Reduce concurrent scraping operations
- Customize user agents to appear more legitimate

---

## Comment Retrieval Failures

**Symptoms**:
- Some comments missing from post responses
- "More comments" placeholders not expanding
- Incomplete comment threads

**Possible Causes**:
- Reddit API "more comments" expansion limits
- Very large comment trees exceeding processing capacity
- API changes in comment retrieval endpoints

**Diagnostic Steps**:
1. Check logs for "more comments" processing:
   ```bash
   grep "more comment" /var/log/reddit-ingestion.log
   ```

2. Compare retrieved comment count with visible count on Reddit:
   ```bash
   grep "comments scraped" /var/log/reddit-ingestion.log
   ```

3. Watch for API errors specific to comment expansion:
   ```bash
   grep "morechildren" /var/log/reddit-ingestion.log
   ```

**Solutions**:
- Increase `MAX_ITERATIONS` in the comment expansion code
- Adjust worker count for parallel comment fetching
- Add delay between comment expansion requests
- For very large posts, break retrieval into smaller chunks

---

## Memory Consumption Issues

**Symptoms**:
- Service crashes with OOM (Out of Memory) errors
- Container restarts frequently
- High memory usage metrics
- Slow garbage collection

**Possible Causes**:
- Scraping very large subreddits or posts
- Memory leaks in comment tree processing
- Excessive concurrency
- Large response bodies not being properly handled

**Diagnostic Steps**:
1. Monitor memory usage during large requests:
   ```bash
   ps -o pid,rss,command | grep reddit-ingestion
   ```

2. Check which endpoints consume most memory:
   ```bash
   grep "memory allocation" /var/log/reddit-ingestion.log
   ```

3. Look for memory growth patterns in monitoring.

**Solutions**:
- Limit maximum comments per post with query parameters
- Add memory limits to container configuration
- Implement pagination for large data sets
- Reduce parallel processing for memory-intensive operations
- Add memory profiling in development

---

## Slow Response Times

**Symptoms**:
- API requests take >5s to complete
- Timeouts from client applications
- High latency metrics

**Possible Causes**:
- Slow proxies
- Reddit API slowdowns
- Too many concurrent requests
- Inefficient comment tree processing

**Diagnostic Steps**:
1. Check response times in logs:
   ```bash
   grep "response time" /var/log/reddit-ingestion.log
   ```

2. Measure proxy performance:
   ```bash
   time curl -x http://username:password@host:port https://old.reddit.com
   ```

3. Monitor Reddit's status for known issues.

**Solutions**:
- Use faster proxies
- Implement request caching for frequent queries
- Optimize comment tree traversal logic
- Reduce depth of comment retrieval for large posts
- Scale service horizontally

---

## Parser Errors

**Symptoms**:
- Responses with missing or null fields
- Errors containing "parse error" or "unmarshal"
- Inconsistent data structures

**Possible Causes**:
- Reddit API response format changes
- Inconsistent Reddit API responses
- JSON parsing errors

**Diagnostic Steps**:
1. Check for JSON parsing errors:
   ```bash
   grep "unmarshal" /var/log/reddit-ingestion.log
   ```

2. Look for null fields in the responses:
   ```bash
   grep "nil pointer" /var/log/reddit-ingestion.log
   ```

3. Compare current response structure with parser expectations.

**Solutions**:
- Update parser code to handle changed fields
- Add null checks in parser logic
- Make field parsing more flexible with optional fields
- Log raw responses for debugging
- Run parser tests to verify fix: `go test ./testing/parser`

---

## API Usage Problems

**Symptoms**:
- HTTP 400 (Bad Request) responses
- Missing or invalid parameter errors
- Incorrect result format

**Possible Causes**:
- Missing required parameters
- Invalid parameter values
- Incorrect endpoint usage

**Diagnostic Steps**:
1. Check API usage in client applications
2. Review parameter validation errors in logs
3. Verify endpoint documentation matches implementation

**Solutions**:
- Refer to [getting-started.md](./getting-started.md) for correct API usage
- Ensure all required parameters are provided
- Validate parameter values before sending requests
- Test requests with curl or Postman before integration
- Run API tests to verify correct behavior: `go test ./testing/api`

---

## Testing Issues

**Symptoms**:
- Test failures in CI/CD pipeline
- Inconsistent test results
- Integration tests timing out

**Possible Causes**:
- Missing test fixtures
- URL path mismatches in mock client
- Environment variable issues
- Test timeout configuration

**Diagnostic Steps**:
1. Run tests with verbose output:
   ```bash
   go test -v ./testing/...
   ```

2. Check for specific test errors:
   ```bash
   go test -v ./testing/integration 2>&1 | grep "FAIL"
   ```

3. Verify mock URL matching in integration tests

**Solutions**:
- For fixture errors: Ensure test runs have proper permissions
- For timeouts: Increase test timeout duration
- For URL mismatches: Use partial URL matching in mocks
- For HTTP errors: Check mock response format matches parser expectations
- Run specific test files when debugging: `go test ./testing/scraper/scraper_test.go`

---

## Related Docs

- [Runbook](./runbook.md)
- [Observability](./observability.md)
- [Getting Started](./getting-started.md)