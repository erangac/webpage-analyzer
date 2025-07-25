package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// httpClient implements the HTTPClient interface.
type httpClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client instance.
func NewHTTPClient() HTTPClient {
	return &httpClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableCompression: false,
				DisableKeepAlives:  false,
			},
		},
	}
}

// FetchWebpage fetches a webpage and returns its content, status code, and any error.
func (c *httpClient) FetchWebpage(ctx context.Context, urlStr string) ([]byte, int, error) {
	// Validate URL format first.
	if err := c.validateURL(urlStr); err != nil {
		return nil, 400, fmt.Errorf("invalid URL format: %v", err)
	}

	// Create request with proper headers.
	httpReq, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, 400, fmt.Errorf("failed to create request: %v", err)
	}

	// Add proper headers.
	httpReq.Header.Set("User-Agent", "WebpageAnalyzer/1.0")
	httpReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	httpReq.Header.Set("Accept-Language", "en-US,en;q=0.5")
	httpReq.Header.Set("Accept-Encoding", "gzip, deflate")
	httpReq.Header.Set("Connection", "keep-alive")

	// Fetch the webpage.
	resp, err := c.client.Do(httpReq)
	if err != nil {
		// Categorize network errors and provide appropriate status codes.
		statusCode, errorMsg := c.categorizeNetworkError(err, urlStr)
		return nil, statusCode, fmt.Errorf(errorMsg)
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, resp.StatusCode, nil
}

// validateURL checks if the URL is properly formatted.
func (c *httpClient) validateURL(urlStr string) error {
	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("malformed URL: %v", err)
	}
	return nil
}

// categorizeNetworkError categorizes network errors and returns appropriate status codes and messages.
func (c *httpClient) categorizeNetworkError(err error, urlStr string) (int, string) {
	errStr := err.Error()

	// DNS resolution errors.
	if strings.Contains(errStr, "no such host") || strings.Contains(errStr, "lookup") {
		return 404, "DNS resolution failed: The domain could not be found. Please check if the URL is correct."
	}

	// Connection refused/timeout errors.
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "connect: connection refused") {
		return 503, "Connection refused: The server is not accepting connections. The service might be down or the port might be closed."
	}

	// Network timeout errors.
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded") {
		return 408, "Request timeout: The server took too long to respond. Please try again later."
	}

	// SSL/TLS errors.
	if strings.Contains(errStr, "certificate") || strings.Contains(errStr, "tls") || strings.Contains(errStr, "ssl") {
		return 495, "SSL/TLS error: There was a problem with the security certificate. The connection is not secure."
	}

	// Protocol errors.
	if strings.Contains(errStr, "protocol") || strings.Contains(errStr, "unsupported protocol") {
		return 400, "Protocol error: The URL uses an unsupported protocol. Please use http:// or https://."
	}

	// Network unreachable.
	if strings.Contains(errStr, "network is unreachable") || strings.Contains(errStr, "no route to host") {
		return 503, "Network unreachable: Cannot reach the server. Please check your internet connection."
	}

	// Generic network error.
	return 503, fmt.Sprintf("Network error: %v. Please check your internet connection and try again.", err)
}

// ParseHTML parses HTML content and returns the document node.
func (c *httpClient) ParseHTML(content []byte) (interface{}, error) {
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}
	return doc, nil
}
