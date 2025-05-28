# Runbook — `reddit-ingestion`

Last updated: 2025-04-17

This runbook defines standard operating procedures and emergency playbooks for the `reddit-ingestion` service. Use it when operating the service in production environments or when troubleshooting issues.

---

## Service Summary

| Field              | Value                                                    |
|--------------------|----------------------------------------------------------|
| Service Name       | `reddit-ingestion`                                       |
| Deployment Targets | production, staging, dev                                 |

---

## Health Checks & Monitoring

| Check                      | Endpoint/Method                      | Expected Behavior                     |
|----------------------------|--------------------------------------|----------------------------------------|
| API Health Check           | `GET /health`                        | Returns `200 OK` status code          |
| Proxy Connectivity         | Check logs for "proxy failed" patterns | No consistent proxy failures         |
| Reddit API Status          | Monitor 429/403 error rates          | Low error rate (< 5%)                |
| Memory Usage               | Container/host metrics               | < 80% of allocated memory            |
| Request Latency            | Endpoint response time metrics       | p95 < 2s for most endpoints          |

---

## Operational Commands

### Production Environment (Docker)

| Action                             | Command                                         |
|------------------------------------|--------------------------------------------------|
| Check service logs                 | `docker logs -f reddit-ingestion`                |
| Restart service                    | `docker restart reddit-ingestion`                |
| Update and rebuild                 | `docker-compose up -d --build`                   |
| Scale with Docker Swarm (if used)  | `docker service scale reddit-ingestion_app=3`    |
| Check container health             | `docker inspect --format "{{.State.Health.Status}}" reddit-ingestion` |
| Update proxies                     | Update .env file and restart container           |

### Docker Environment

| Action                             | Command                                         |
|------------------------------------|--------------------------------------------------|
| Check service logs                 | `docker logs -f reddit-ingestion`                |
| Restart service                    | `docker restart reddit-ingestion`                |
| Update and rebuild                 | `docker-compose up -d --build`                   |
| Check proxy status                 | `docker exec -it reddit-ingestion curl -x [proxy-url] https://httpbin.org/ip` |

---

## Common Incidents & Response

### 1. Service Not Responding

**Symptoms**: 
- `/health` endpoint returns non-200 response
- Client requests timeout

**Response**:
1. Check service logs for errors: `kubectl logs -l app=reddit-ingestion -f`
2. Verify proxy connectivity is working
3. Check for memory or CPU constraints
4. Restart the service if necessary: `kubectl rollout restart deployment reddit-ingestion`
5. If issue persists, check for Reddit API changes or proxy blocks

---

### 2. High Error Rate from Reddit API

**Symptoms**:
- Logs show many 429 (Too Many Requests) or 403 (Forbidden) responses
- Data retrieval is incomplete

**Response**:
1. Verify proxy URLs are valid and working
2. Check if Reddit is experiencing issues (status.reddit.com)
3. Add more proxies to the rotation
4. Increase `PROXY_MAX_RETRIES` and `RATE_LIMIT_DELAY`
5. Consider implementing a circuit breaker to pause requests temporarily

---

### 3. Memory Usage Spikes

**Symptoms**:
- Container memory usage approaches limits
- OOM (Out of Memory) errors in logs

**Response**:
1. Check which endpoints are being called (large subreddits or posts with many comments)
2. Limit concurrent requests or reduce page size limits
3. Increase memory allocation if needed
4. Implement pagination for large result sets
5. Investigate possible memory leaks if issue persists

---

### 4. Slow Response Times

**Symptoms**:
- Endpoints take >30s to respond
- Post Endpoint can take upto 300s to respond depending upon the no of comments
- Timeouts occur frequently

**Response**:
1. Check Reddit API response times in logs
2. Verify proxy speed and connectivity
3. Monitor CPU usage for potential bottlenecks
4. Consider scaling up the service horizontally
5. Implement caching for frequently accessed data

---

## Related Docs

- [Observability](./observability.md)
- [Deployment](./deployment.md)
- [Troubleshooting](./troubleshooting.md)