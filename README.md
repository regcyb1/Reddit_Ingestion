# Reddit Ingestion System

Go service for ingesting Reddit content through proxies to avoid rate limiting.

## Purpose

This service provides a reliable way to extract data from Reddit via HTTP endpoints. It handles proxy rotation, browser fingerprinting, and implements retries to ensure consistent data collection despite Reddit's rate limits.

## Features

- **Proxy Rotation**: Cycles through proxies to avoid rate limits
- **Browser Fingerprinting**: Mimics real browsers to avoid detection
- **Configurable Endpoints**: Extract subreddit posts, user data, posts, and search results
- **Reliable Comment Scraping**: Handles Reddit's nested comment structure
- **Resilient Error Handling**: Implements retries with backoff
- **Comprehensive Testing**: Unit and integration tests for all components
- **Interactive API Documentation**: Provides Swagger UI for exploring all endpoints

## API Endpoints

| Endpoint       | Description                               |
|----------------|-------------------------------------------|
| `/subreddit`   | Fetch posts from a specific subreddit     |
| `/user`        | Get user information and activity         |
| `/post`        | Get a post with all comments              |
| `/search`      | Search Reddit content with filters        |

## Getting Started

See [Getting Started](./docs/getting-started.md) for setup instructions.

## Documentation

- [Overview](./docs/overview.md)
- [Configuration](./docs/configuration.md)
- [Deployment](./docs/deployment.md)
- [Service Requirement Document](./docs/srd.md)
- [Runbook](./docs/runbook.md)
- [Troubleshooting](./docs/troubleshooting.md)
- [Usage](./docs/usage.md)
- **API Reference**: Available via Swagger UI at `/swagger/index.html`

## Testing

The service includes a comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run integration tests with verbose output
go test -v ./testing/integration

# Run specific component tests
go test ./testing/parser
go test ./testing/scraper
go test ./testing/api
```

## Architecture

The service is built with a clean architecture pattern:

- `internal/client`: API client for Reddit with proxy handling
- `internal/config`: Configuration management
- `internal/parser`: Processing Reddit API responses
- `internal/scraper`: Core scraping functionality
- `pkg/utils`: Shared utilities including proxy rotation and TLS fingerprinting
- `testing`: Test suite with mocks and integration tests

## Project Structure

```
reddit-ingestion/
├── cmd/
│   └── server/
│       └── main.go                  # Application entry point
├── docs/                            # Documentation
|   ├── .md                          # General documentations
│   ├── docs.go                      # Generated Swagger API definitions
│   ├── swagger.json                 # Generated Swagger spec (JSON)
│   └── swagger.yaml                 # Generated Swagger spec (YAML)
├── internal/
│   ├── app/
│   │   └── app.go                   # Application initialization
│   ├── client/
│   │   ├── interface.go             # Client interface definition
│   │   └── reddit_client.go         # Reddit API client implementation
│   ├── config/
│   │   └── config.go                # Configuration loading/management
│   ├── handler/
│   │   └── http/                    # HTTP handlers for API endpoints
│   │       ├── post_handler.go      # Post endpoint handler
│   │       ├── search_handler.go    # Search endpoint handler
│   │       ├── subreddit_handler.go # Subreddit endpoint handler
│   │       └── user_handler.go      # User endpoint handler
│   ├── models/
│   │   ├── models.go                # Data models for Reddit content
│   │   ├── swagger_models.go        # Additional models for Swagger endpoints
│   │   └── swagger_responses.go     # Additional models for Swagger docs
│   ├── parser/
│   │   ├── interface.go             # Parser interface definition
│   │   └── parser.go                # Reddit API response parser
│   ├── router/
│   │   └── router.go                # HTTP router setup
│   └── scraper/
│       └── service.go               # Core scraping functionality
├── pkg/
│   └── utils/
│       └── proxy_client.go          # Proxy rotation and TLS fingerprinting
├── testing/                         # Test suite
│   ├── api/                         # API endpoint tests
│   │   └── api_test.go              # Tests for HTTP handlers
│   ├── integration/                 # Integration tests
│   │   ├── fingerprinting_test.go   
│   │   └── integration_test.go      # End-to-end tests with mocked Reddit API
│   ├── mocks/                       # Mock implementations
│   │   ├── client_mock.go           # Mock Reddit client
│   │   └── parser_mock.go           # Mock parser
│   ├── parser/                      # Parser tests
│   │   └── parser_test.go           # Tests for Reddit API response parsing
│   └── scraper/                     # Scraper tests
│       └── scraper_test.go          # Tests for scraper service
├── .env.example                     # Example environment variables
├── .gitignore                       # Git ignore file
├── Dockerfile                       # Docker build instructions
├── docs.go                          # Root-level Swagger documentation file
├── go.mod                           # Go module definition
├── go.sum                           # Go module checksums
└── README.md                        # This file
```