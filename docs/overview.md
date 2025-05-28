# Overview — `reddit-ingestion`

The `reddit-ingestion` is a data collection service that fetches content from Reddit via their public API. It uses multiple proxies with browser fingerprinting to avoid rate limits and provide reliable access to Reddit data.

For detailed specifications and responsibilities, see the [Service Requirement Document](./srd.md).

---

## What It Does

- Fetches posts from subreddits with customizable limits and filters
- Retrieves user information, posts, and comments
- Gets detailed post data including all comments (even expanding "load more" sections)
- Performs Reddit searches with various filters
- Rotates through multiple proxies to avoid rate limiting
- Implements browser fingerprinting to mimic legitimate browsers
- Retries failed requests with exponential backoff

---

## How It Fits

The system is designed with distinct components:

- **API Layer**: HTTP endpoints for different Reddit content types
- **Scraper Service**: Business logic for data collection with pagination support
- **Parser**: Converts Reddit API responses into structured data models
- **Client**: Handles HTTP requests with proxy rotation and retries
- **Utils**: Proxy management and browser fingerprinting
- **Testing Layer**: Validates system components independently and together

---

## Key Features

1. **Proxy Rotation**: Automatically cycles through multiple proxies to avoid detection
2. **Browser Fingerprinting**: Uses uTLS to mimic Chrome, Firefox, Safari, or Edge browser fingerprints
3. **Resilient Retries**: Implements exponential backoff for failed requests
4. **Comprehensive Comment Scraping**: Efficiently handles Reddit's complex comment pagination
5. **Configurable Limits**: Supports various modes from quick sampling to exhaustive data collection
6. **Interactive API Documentation**: Provides Swagger UI for exploring and testing endpoints
7. **Thorough Test Coverage**: Unit and integration tests for all components

---

## Technologies Used

- **Go**: Primary programming language
- **Echo**: HTTP framework for the API layer
- **uTLS**: For browser fingerprinting
- **Context**: For request cancellation and timeouts
- **Goroutines**: For concurrent processing of data

---

## Common Use Cases

- Monitoring subreddits for new content
- Analyzing user posting behavior
- Retrieving complete comment trees for discussion analysis
- Searching Reddit for specific keywords or phrases

---

## When to Use This Doc

Use this document to:
- Get a high-level understanding of the system
- Understand the main components and how they interact
- Find pointers to more detailed documentation

For details about specific aspects, refer to:
- [Getting Started](./getting-started.md) for setup instructions
- [Configuration](./configuration.md) for environment variables
- [Usage](./usage.md) for API examples
- [SRD](./srd.md) for detailed specifications