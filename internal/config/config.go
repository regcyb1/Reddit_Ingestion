// internal/config/config.go
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ProxyURLs           []string
	UserAgent           string
	MaxRetries          int
	DefaultPostLimit    int
	DefaultCommentLimit int
	ServerPort          string
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	RedditBaseURL       string
	RequestTimeout      time.Duration
	RateLimitDelay      time.Duration
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	proxyURLsStr := os.Getenv("REDDIT_PROXY_URLS")
	var proxyURLs []string

	if proxyURLsStr != "" {
		proxyURLsStr = strings.TrimSpace(proxyURLsStr)

		rawProxies := strings.Split(proxyURLsStr, ",")
		for _, proxy := range rawProxies {
			proxy = strings.TrimSpace(proxy)
			if proxy == "" {
				continue
			}

			if !strings.HasPrefix(proxy, "http://") && !strings.HasPrefix(proxy, "https://") {
				return nil, fmt.Errorf("invalid proxy URL format, must start with http:// or https://: %s", proxy)
			}

			_, err := url.Parse(proxy)
			if err != nil {
				return nil, fmt.Errorf("invalid proxy URL %s: %w", proxy, err)
			}

			proxyURLs = append(proxyURLs, proxy)
		}

		fmt.Printf("Loaded %d proxy URLs from configuration\n", len(proxyURLs))
	}

	if len(proxyURLs) == 0 {
		return nil, fmt.Errorf("REDDIT_PROXY_URLS environment variable is required and must contain at least one valid proxy URL")
	}

	userAgent := os.Getenv("REDDIT_USER_AGENT")
	if userAgent == "" {
		userAgent = "Mozilla/5.0"
		fmt.Println("No user agent specified, using default:", userAgent)
	}

	return &Config{
		ProxyURLs:           proxyURLs,
		UserAgent:           userAgent,
		MaxRetries:          getEnvInt("PROXY_MAX_RETRIES", 3),
		DefaultPostLimit:    getEnvInt("SCRAPER_DEFAULT_POST_LIMIT", 25),
		DefaultCommentLimit: getEnvInt("SCRAPER_DEFAULT_COMMENT_LIMIT", 50),
		ServerPort:          getEnv("SERVER_PORT", "8080"),
		RequestTimeout:      getEnvDuration("REQUEST_TIMEOUT", 30*time.Second),
		ReadTimeout:         getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second),
		WriteTimeout:        getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second),
		RateLimitDelay:      getEnvDuration("RATE_LIMIT_DELAY", 100*time.Millisecond),
		RedditBaseURL:       getEnv("REDDIT_BASE_URL", "https://old.reddit.com"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}
