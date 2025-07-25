package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// httpClient implements the HTTPClient interface
type httpClient struct {
	client *http.Client
}

// NewHTTPClient creates a new HTTP client instance
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

// FetchWebpage fetches a webpage and returns its content, status code, and any error
func (c *httpClient) FetchWebpage(ctx context.Context, url string) ([]byte, int, error) {
	// Create request with proper headers
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %v", err)
	}

	// Add proper headers
	httpReq.Header.Set("User-Agent", "WebpageAnalyzer/1.0")
	httpReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	httpReq.Header.Set("Accept-Language", "en-US,en;q=0.5")
	httpReq.Header.Set("Accept-Encoding", "gzip, deflate")
	httpReq.Header.Set("Connection", "keep-alive")

	// Fetch the webpage
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, resp.StatusCode, nil
}

// ParseHTML parses HTML content and returns the document node
func (c *httpClient) ParseHTML(content []byte) (interface{}, error) {
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}
	return doc, nil
} 