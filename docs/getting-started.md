# Getting Started — `reddit-ingestion`

This guide explains how to set up my Reddit ingestion service.

---

## Requirements

- Go >= 1.24.2
- At least one HTTP/HTTPS proxy (required)
- Git

---

## Quickstart

```bash
# Clone the repository
git clone https://github.com/your-username/reddit-ingestion.git
cd reddit-ingestion

# Copy example environment file
cp .env.example .env

# Edit the .env file to add your proxy URLs and customize settings
# IMPORTANT: You must add at least one proxy URL

# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go

# Access Swagger UI in your browser
open http://localhost:8080/swagger/index.html
Note: replace localhost with your VDI IP
```

---

## Development Commands

| Command                              | Description                            |
|--------------------------------------|----------------------------------------|
| `go run cmd/server/main.go`          | Start the server                       |
| `go test ./...`                      | Run all tests                          |
| `go test -v ./testing/integration`   | Run integration tests with verbose output |
| `go test ./testing/parser`           | Run parser component tests             |
| `go test ./testing/scraper`          | Run scraper component tests            |
| `go test ./testing/api`              | Run API handler tests                  |
| `go build -o reddit-ingestion ./cmd/server/main.go` | Build the binary        |
| `docker build -t reddit-ingestion .` | Build Docker image                     |

---

## Setting Up Proxies

The system requires at least one HTTP/HTTPS proxy. Add your proxies to the `.env` file:

```
REDDIT_PROXY_URLS=http://username:password@proxy1.example.com:8080,http://username:password@proxy2.example.com:8080
```

---

## API Testing

After starting the server, test the API endpoints:

### Fetch Subreddit Posts

```bash
curl "http://localhost:8080/subreddit?subreddit=golang&limit=5"
```

### Get User Info

```bash
curl "http://localhost:8080/user?username=spez&post_limit=5&comment_limit=5"
```

### Get Post Info

```bash
curl "http://localhost:8080/post?post_id=13fg9z"
```

### Search Reddit

```bash
curl "http://localhost:8080/search?search_string=golang&limit=5&sort=new"
```

---

## Testing the Application

The application includes a comprehensive test suite:

### Unit Tests

Tests for individual components:
- Parser tests verify Reddit JSON response parsing
- Scraper tests validate data collection logic
- API tests ensure proper HTTP handler functionality

### Integration Tests

End-to-end tests that validate the entire system flow:
- Tests each API endpoint with mock Reddit responses
- Verifies proper data transformation
- Confirms system components work together correctly

### Running Tests

Run all tests:
```bash
go test ./...
```

Run integration tests with detailed output:
```bash
go test -v ./testing/integration
```

---

## Common Issues

- **Proxy Connection Errors**: Check that your proxies are working correctly
- **Rate Limiting**: If you see 429 errors, add more proxies or increase delay
- **Memory Usage**: For large subreddits/posts, you might need more memory
- **Test Failures**: If tests fail, ensure your Go version is compatible and environment is properly configured