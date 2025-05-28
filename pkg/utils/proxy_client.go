// pkg/utils/proxy_client.go
package utils

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	utls "github.com/refraction-networking/utls"
	proxy "golang.org/x/net/proxy"
)

type BrowserType int

const (
	Chrome BrowserType = iota
	Firefox
	Safari
	Edge
	EnvUseRandomUserAgents = "USE_RANDOM_USER_AGENTS"
)

var clientHelloIDs = []utls.ClientHelloID{
	utls.HelloChrome_Auto,
	utls.HelloFirefox_Auto,
	utls.HelloSafari_Auto,
	utls.HelloEdge_Auto,
}

var acceptLanguages = []string{
	"en-US,en;q=0.9",
	"en-US,en;q=0.8",
	"en-GB,en;q=0.9,en-US;q=0.8",
	"en-CA,en;q=0.9,fr-CA;q=0.8",
	"fr-FR,fr;q=0.9,en;q=0.8",
	"de-DE,de;q=0.9,en;q=0.8",
	"es-ES,es;q=0.9,en;q=0.8",
	"it-IT,it;q=0.9,en;q=0.8",
}

var userAgents = map[BrowserType][]string{
	Chrome: {
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	},
	Firefox: {
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:124.0) Gecko/20100101 Firefox/124.0",
		"Mozilla/5.0 (X11; Linux x86_64; rv:122.0) Gecko/20100101 Firefox/122.0",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:124.0) Gecko/20100101 Firefox/124.0",
	},
	Safari: {
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
	},
	Edge: {
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.2277.128",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.2365.66",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.2365.80",
	},
}

func shouldUseRandomUserAgents() bool {
	useRandomStr := os.Getenv(EnvUseRandomUserAgents)
	if useRandomStr != "" {
		useRandom, err := strconv.ParseBool(useRandomStr)
		if err == nil {
			return useRandom
		}
	}

	return true
}

func getCorrespondingBrowserType(clientHelloID utls.ClientHelloID) BrowserType {
	switch clientHelloID {
	case utls.HelloFirefox_Auto:
		return Firefox
	case utls.HelloSafari_Auto:
		return Safari
	case utls.HelloEdge_Auto:
		return Edge
	default:
		return Chrome
	}
}

func randomItem[T any](items []T) T {
	return items[rand.Intn(len(items))]
}

func addRandomizedBrowserHeaders(req *http.Request, browserType BrowserType, userAgent string) {
	if !shouldUseRandomUserAgents() && userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	} else {
		req.Header.Set("User-Agent", randomItem(userAgents[browserType]))
	}

	req.Header.Set("Accept-Language", randomItem(acceptLanguages))

	encodings := []string{
		"gzip, deflate, br",
		"gzip, deflate",
		"br, gzip, deflate",
	}
	req.Header.Set("Accept-Encoding", randomItem(encodings))

	cacheControls := []string{
		"max-age=0",
		"no-cache",
		"max-age=0, private, must-revalidate",
	}
	if rand.Intn(10) > 2 {
		req.Header.Set("Cache-Control", randomItem(cacheControls))
	}

	if rand.Intn(10) > 3 {
		req.Header.Set("DNT", fmt.Sprintf("%d", rand.Intn(2)+1))
	}

	switch browserType {
	case Chrome, Edge:
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")

		if rand.Intn(10) > 2 {
			req.Header.Set("Sec-Fetch-Dest", "document")
			req.Header.Set("Sec-Fetch-Mode", "navigate")
			req.Header.Set("Sec-Fetch-Site", "none")

			if rand.Intn(10) > 3 {
				req.Header.Set("Sec-Fetch-User", "?1")
			}
		}

	case Firefox:
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")

		if rand.Intn(10) > 2 {
			req.Header.Set("TE", "trailers")
		}

	case Safari:
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		if rand.Intn(10) > 2 {
			req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		}
	}

	if rand.Intn(10) > 0 {
		req.Header.Set("Upgrade-Insecure-Requests", "1")
	}

	connections := []string{"keep-alive", "close"}
	req.Header.Set("Connection", randomItem(connections))
}

type ProxyRotator struct {
	proxyURLs  []string
	parsedURLs []*url.URL
	currentIdx uint32
	mutex      sync.RWMutex
}

func NewProxyRotator(proxyURLs []string) (*ProxyRotator, error) {
	rotator := &ProxyRotator{
		proxyURLs:  proxyURLs,
		currentIdx: 0,
	}

	for _, rawURL := range proxyURLs {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse proxy URL %s: %w", rawURL, err)
		}
		rotator.parsedURLs = append(rotator.parsedURLs, parsedURL)
	}

	return rotator, nil
}

func (r *ProxyRotator) NextProxy() *url.URL {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.parsedURLs) == 0 {
		return nil
	}

	idx := atomic.AddUint32(&r.currentIdx, 1) % uint32(len(r.parsedURLs))
	return r.parsedURLs[idx]
}

func (r *ProxyRotator) GetProxyForID(id uint32) *url.URL {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if len(r.parsedURLs) == 0 {
		return nil
	}

	idx := id % uint32(len(r.parsedURLs))
	return r.parsedURLs[idx]
}

type FingerprintingDialer struct {
	proxyURL      *url.URL
	clientHelloID utls.ClientHelloID
	browserType   BrowserType
}

func NewFingerprintingDialer(proxyURL *url.URL) *FingerprintingDialer {
	helloID := clientHelloIDs[rand.Intn(len(clientHelloIDs))]
	browserType := getCorrespondingBrowserType(helloID)

	return &FingerprintingDialer{
		proxyURL:      proxyURL,
		clientHelloID: helloID,
		browserType:   browserType,
	}
}

func (d *FingerprintingDialer) DialTLSContext(ctx context.Context, network, addr string) (net.Conn, error) {
	var conn net.Conn
	var err error

	if d.proxyURL == nil {
		var dialer net.Dialer
		conn, err = dialer.DialContext(ctx, network, addr)
		if err != nil {
			return nil, fmt.Errorf("direct dial: %w", err)
		}
	} else {
		conn, err = d.dialThroughProxyWithContext(ctx, network, addr)
		if err != nil {
			return nil, fmt.Errorf("proxy dial: %w", err)
		}
	}

	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}

	config := &utls.Config{
		ServerName: host,
	}

	uconn := utls.UClient(conn, config, d.clientHelloID)
	if err := uconn.Handshake(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("uTLS handshake: %w", err)
	}

	return uconn, nil
}

func (d *FingerprintingDialer) dialThroughProxyWithContext(ctx context.Context, network, addr string) (net.Conn, error) {
	switch d.proxyURL.Scheme {
	case "http", "https":
		transport := &http.Transport{
			Proxy: http.ProxyURL(d.proxyURL),
		}

		req, err := http.NewRequestWithContext(ctx, "CONNECT", "https://"+addr, nil)
		if err != nil {
			return nil, fmt.Errorf("create CONNECT request: %w", err)
		}

		if d.proxyURL.User != nil {
			if password, ok := d.proxyURL.User.Password(); ok {
				req.SetBasicAuth(d.proxyURL.User.Username(), password)
			}
		}

		conn, err := transport.DialContext(ctx, network, addr)
		if err != nil {
			return nil, fmt.Errorf("dial via HTTP proxy: %w", err)
		}

		return conn, nil

	case "socks5":
		auth := &proxy.Auth{}
		if d.proxyURL.User != nil {
			auth.User = d.proxyURL.User.Username()
			if password, ok := d.proxyURL.User.Password(); ok {
				auth.Password = password
			}
		}

		dialer, err := proxy.SOCKS5("tcp", d.proxyURL.Host, auth, &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		})
		if err != nil {
			return nil, fmt.Errorf("create SOCKS5 dialer: %w", err)
		}

		connCh := make(chan net.Conn, 1)
		errCh := make(chan error, 1)

		go func() {
			conn, err := dialer.Dial(network, addr)
			if err != nil {
				errCh <- err
				return
			}
			connCh <- conn
		}()

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case conn := <-connCh:
			return conn, nil
		case err := <-errCh:
			return nil, fmt.Errorf("dial via SOCKS5 proxy: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", d.proxyURL.Scheme)
	}
}

type TLSFingerprintingTransport struct {
	proxyRotator *ProxyRotator
	transport    *http.Transport
}

func NewTLSFingerprintingTransport(rotator *ProxyRotator) http.RoundTripper {
	transport := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ForceAttemptHTTP2:     false,
		DisableCompression:    false,
	}

	return &TLSFingerprintingTransport{
		proxyRotator: rotator,
		transport:    transport,
	}
}

func (t *TLSFingerprintingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqCopy := req.Clone(req.Context())

	existingUserAgent := req.Header.Get("User-Agent")

	goroutineID := uint32(time.Now().UnixNano())
	proxyURL := t.proxyRotator.GetProxyForID(goroutineID)

	var browserType BrowserType

	if proxyURL != nil {
		t.transport.Proxy = http.ProxyURL(proxyURL)
	}

	if req.URL.Scheme == "https" {
		dialer := NewFingerprintingDialer(proxyURL)
		t.transport.DialTLSContext = dialer.DialTLSContext
		browserType = dialer.browserType
	} else {
		browserType = BrowserType(rand.Intn(4))
	}

	addRandomizedBrowserHeaders(reqCopy, browserType, existingUserAgent)

	return t.transport.RoundTrip(reqCopy)
}

func maskProxyURL(proxyURL string) string {
	if !strings.Contains(proxyURL, "@") {
		return proxyURL
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		parts := strings.Split(proxyURL, "@")
		if len(parts) > 1 {
			auth := strings.Split(parts[0], "://")
			protocol := ""
			if len(auth) > 1 {
				protocol = auth[0] + "://"
				auth[0] = auth[1]
			}

			userPass := strings.Split(auth[0], ":")
			if len(userPass) > 1 {
				return protocol + userPass[0] + ":****@" + parts[1]
			}
		}
		return "[masked]"
	}

	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		return strings.Replace(proxyURL, parsedURL.User.String(), username+":****", 1)
	}

	return proxyURL
}

type RetryableClient struct {
	client     *http.Client
	maxRetries int
	userAgent  string
}

func NewRetryableClient(proxyURLs []string, maxRetries int, userAgent string) (*RetryableClient, error) {
	if len(proxyURLs) == 0 {
		return nil, fmt.Errorf("at least one proxy URL must be provided")
	}

	var validProxies []string
	for _, proxy := range proxyURLs {
		if proxy != "" {
			validProxies = append(validProxies, proxy)
		}
	}

	if len(validProxies) == 0 {
		return nil, fmt.Errorf("no valid proxy URLs provided")
	}

	for i, proxy := range validProxies {
		maskedProxy := maskProxyURL(proxy)
		fmt.Printf("Proxy #%d: %s\n", i+1, maskedProxy)
	}

	rotator, err := NewProxyRotator(validProxies)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy rotator: %w", err)
	}

	httpClient := &http.Client{
		Transport: NewTLSFingerprintingTransport(rotator),
		Timeout:   30 * time.Second,
	}

	fmt.Printf("Created HTTP client with %d proxies and TLS fingerprinting\n", len(validProxies))

	return &RetryableClient{
		client:     httpClient,
		maxRetries: maxRetries,
		userAgent:  userAgent,
	}, nil
}

func (c *RetryableClient) Do(req *http.Request) (*http.Response, []byte, error) {
	var resp *http.Response
	var bodyBytes []byte
	var err error

	if req.Header.Get("User-Agent") == "" && !shouldUseRandomUserAgents() {
		req.Header.Set("User-Agent", c.userAgent)
	}

	var reqBody []byte
	if req.Body != nil {
		reqBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("reading request body: %w", err)
		}
		req.Body.Close()
	}

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if reqBody != nil {
			req.Body = io.NopCloser(bytes.NewReader(reqBody))
		}

		if attempt > 0 {
			backoffTime := time.Duration(1<<uint(attempt)) * time.Second
			time.Sleep(backoffTime)

			fmt.Printf("Retry attempt %d after waiting %v\n", attempt+1, backoffTime)
		}

		resp, err = c.client.Do(req)
		if err != nil {
			fmt.Printf("Request error (attempt %d): %v\n", attempt+1, err)

			if attempt == c.maxRetries-1 {
				return nil, nil, fmt.Errorf("all %d attempts failed: %w", c.maxRetries, err)
			}
			continue
		}

		var reader io.ReadCloser
		switch strings.ToLower(resp.Header.Get("Content-Encoding")) {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				resp.Body.Close()
				fmt.Printf("Error creating gzip reader (attempt %d): %v\n", attempt+1, err)
				if attempt == c.maxRetries-1 {
					return nil, nil, fmt.Errorf("failed to decompress gzip response: %w", err)
				}
				continue
			}
			defer reader.Close()
		default:
			reader = resp.Body
		}

		bodyBytes, err = io.ReadAll(reader)
		resp.Body.Close()

		if err != nil {
			fmt.Printf("Error reading response body (attempt %d): %v\n", attempt+1, err)

			if attempt == c.maxRetries-1 {
				return nil, nil, fmt.Errorf("reading response body: %w", err)
			}
			continue
		}

		if len(bodyBytes) > 0 && bodyBytes[0] == 0x1f && bodyBytes[1] == 0x8b {
			gr, err := gzip.NewReader(bytes.NewReader(bodyBytes))
			if err == nil {
				uncompressed, err := io.ReadAll(gr)
				gr.Close()
				if err == nil {
					fmt.Printf("Detected and uncompressed double-gzipped content\n")
					bodyBytes = uncompressed
				}
			}
		}

		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			fmt.Printf("Received status code %d (attempt %d)\n", resp.StatusCode, attempt+1)

			if attempt == c.maxRetries-1 {
				return nil, nil, fmt.Errorf("server error: status %d", resp.StatusCode)
			}
			continue
		}

		break
	}

	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return resp, bodyBytes, nil
}
