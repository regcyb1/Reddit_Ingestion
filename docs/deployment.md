# Deployment — `reddit-ingestion`

This document describes how to deploy the `reddit-ingestion` service.

---

## Deployment Model

| Environment | Description                                | Deployment Method                              |
|-------------|--------------------------------------------|------------------------------------------------|
| Local       | Development environment                    | `go run cmd/server/main.go`                    |
| Docker      | Production deployment                      | Docker container                               |

---

## Deployment Notes

- The service is **stateless** - I can restart it without data loss
- It requires:
  - Valid HTTP/HTTPS proxies for Reddit access
  - Internet connection to reach Reddit
  - Environment variables (see Configuration doc)

---

## Docker Deployment

I've created a simple Dockerfile for the service:

```dockerfile
FROM golang:1.24.2 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./server ./cmd/server

CMD ["./server"]
```

To build and run the container:

```bash
# Build the image
docker build -t reddit-ingestion .

# Run the container
docker run -d -p 8080:8080 --env-file .env --name reddit-ingestion reddit-ingestion
```

---

## Special Production Considerations

- **Proxy Management**: Production should use a robust pool of proxies with automatic rotation
- **Rate Limiting**: Configure appropriate delays between requests to avoid proxy bans
- **Resource Scaling**: Monitor CPU/memory usage to scale appropriately
- **Secrets Management**: Store proxy credentials and API keys securely in Docker secrets or environment files
- **Container Orchestration**: Consider using Docker Swarm for simple orchestration if needed
- **Backup Environment**: Maintain a standby environment in case of issues with the main deployment

---

## Rollback Procedure

1. Identify the last stable version
2. Stop the current container: `docker stop reddit-ingestion`
3. Start the container with the previous image version: `docker run -d --name reddit-ingestion -p 8080:8080 --env-file .env.production reddit-ingestion:previous-version`
4. If using Docker Compose: Update the image tag in docker-compose.yml and run `docker-compose up -d`
5. Monitor logs and metrics to ensure proper function

---

## Required Configuration

All runtime config is managed via environment variables. See [Configuration](./configuration.md) for full list.

Minimum required variables for production:
- `REDDIT_PROXY_URLS` (at least 5 proxies recommended)
- `REDDIT_USER_AGENT`
- `SERVER_PORT`

---

## Health Checks

The service exposes a health endpoint at `/health` that returns a 200 OK response when the service is running normally.

---

## Related Docs

- [Getting Started](./getting-started.md)
- [Configuration](./configuration.md)
- [Runbook](./runbook.md)
- [Observability](./observability.md)
