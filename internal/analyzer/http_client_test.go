package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient()
	require.NotNil(t, client, "NewHTTPClient() should not return nil")
}

func TestHTTPClient_FetchWebpage_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><head><title>Test Page</title></head><body>Hello World</body></html>`))
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, server.URL)

	require.NoError(t, err, "FetchWebpage() should not return error")
	assert.Equal(t, http.StatusOK, statusCode, "Status code should be OK")
	assert.NotEmpty(t, content, "FetchWebpage() should not return empty content")

	// Check if the HTML contains expected content
	htmlStr := string(content)
	assert.Contains(t, htmlStr, "Test Page", "Response should contain expected title")
	assert.Contains(t, htmlStr, "Hello World", "Response should contain expected body")
}

func TestHTTPClient_FetchWebpage_404Error(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, server.URL+"/nonexistent")

	require.NoError(t, err, "FetchWebpage() should not return error for 404")
	assert.Equal(t, http.StatusNotFound, statusCode, "Status code should be 404")
	assert.NotEmpty(t, content, "FetchWebpage() should return content even for 404")
}

func TestHTTPClient_FetchWebpage_500Error(t *testing.T) {
	// Create a test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, server.URL)

	require.NoError(t, err, "FetchWebpage() should not return error for 500")
	assert.Equal(t, http.StatusInternalServerError, statusCode, "Status code should be 500")
	assert.NotEmpty(t, content, "FetchWebpage() should return content even for 500")
}

func TestHTTPClient_FetchWebpage_InvalidURL(t *testing.T) {
	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, "invalid-url")

	require.Error(t, err, "FetchWebpage() should return error for invalid URL")
	assert.Nil(t, content, "FetchWebpage() should return nil content for invalid URL")
	assert.Equal(t, 400, statusCode, "Status code should be 400 for invalid URL")
}

func TestHTTPClient_FetchWebpage_EmptyURL(t *testing.T) {
	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, "")

	require.Error(t, err, "FetchWebpage() should return error for empty URL")
	assert.Nil(t, content, "FetchWebpage() should return nil content for empty URL")
	assert.Equal(t, 400, statusCode, "Status code should be 400 for empty URL")
}

func TestHTTPClient_FetchWebpage_Timeout(t *testing.T) {
	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond) // Short delay
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Delayed Response</body></html>"))
	}))
	defer server.Close()

	// Create a client with very short timeout for testing
	testClient := &httpClient{
		client: &http.Client{
			Timeout: 10 * time.Millisecond, // Very short timeout for testing
		},
	}

	ctx := context.Background()
	content, statusCode, err := testClient.FetchWebpage(ctx, server.URL)

	require.Error(t, err, "FetchWebpage() should return error for timeout")
	assert.Nil(t, content, "FetchWebpage() should return nil content for timeout")
	assert.Equal(t, 408, statusCode, "Status code should be 408 for timeout")
}

func TestHTTPClient_FetchWebpage_NonHTMLContent(t *testing.T) {
	// Create a test server that returns JSON instead of HTML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "This is JSON, not HTML"}`))
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, server.URL)

	require.NoError(t, err, "FetchWebpage() should not return error for non-HTML content")
	assert.Equal(t, http.StatusOK, statusCode, "Status code should be OK")
	assert.NotEmpty(t, content, "FetchWebpage() should return content even if it's not HTML")

	// Should still return the content even if it's not HTML
	contentStr := string(content)
	assert.Contains(t, contentStr, "This is JSON, not HTML", "Should return the content regardless of type")
}

func TestHTTPClient_FetchWebpage_LargeResponse(t *testing.T) {
	// Create a test server that returns a large HTML response
	largeHTML := `<!DOCTYPE html><html><head><title>Large Page</title></head><body>`
	for i := 0; i < 1000; i++ {
		largeHTML += `<p>This is paragraph number ` + string(rune(i)) + `</p>`
	}
	largeHTML += `</body></html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeHTML))
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, server.URL)

	require.NoError(t, err, "FetchWebpage() should not return error for large response")
	assert.Equal(t, http.StatusOK, statusCode, "Status code should be OK")
	assert.NotEmpty(t, content, "FetchWebpage() should not return empty content for large response")

	contentStr := string(content)
	assert.Contains(t, contentStr, "Large Page", "Should contain the title from large response")
	assert.Greater(t, len(contentStr), 1000, "Should return the full large response")
}

func TestHTTPClient_FetchWebpage_Redirect(t *testing.T) {
	// Create a test server that redirects
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusMovedPermanently)
			return
		}
		if r.URL.Path == "/final" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<!DOCTYPE html><html><head><title>Final Page</title></head><body>Redirected Successfully</body></html>`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, server.URL+"/redirect")

	require.NoError(t, err, "FetchWebpage() should not return error for redirect")
	assert.Equal(t, http.StatusOK, statusCode, "Status code should be OK after redirect")
	assert.NotEmpty(t, content, "FetchWebpage() should not return empty content for redirect")

	contentStr := string(content)
	assert.Contains(t, contentStr, "Final Page", "Should follow redirect and return final page content")
}

func TestHTTPClient_FetchWebpage_UserAgent(t *testing.T) {
	var userAgent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userAgent = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html><html><body>Test</body></html>`))
	}))
	defer server.Close()

	client := NewHTTPClient()
	ctx := context.Background()
	_, _, err := client.FetchWebpage(ctx, server.URL)

	require.NoError(t, err, "FetchWebpage() should not return error")
	assert.NotEmpty(t, userAgent, "FetchWebpage() should set User-Agent header")
	assert.Contains(t, userAgent, "WebpageAnalyzer", "User-Agent should contain 'WebpageAnalyzer'")
}

func TestHTTPClient_ParseHTML_Success(t *testing.T) {
	client := NewHTTPClient()
	htmlContent := []byte(`<!DOCTYPE html><html><head><title>Test</title></head><body>Hello</body></html>`)

	doc, err := client.ParseHTML(htmlContent)

	require.NoError(t, err, "ParseHTML() should not return error")
	require.NotNil(t, doc, "ParseHTML() should not return nil document")

	// Check if it's a valid HTML node
	htmlNode, ok := doc.(*html.Node)
	require.True(t, ok, "ParseHTML() should return *html.Node")
	assert.Equal(t, html.DocumentNode, htmlNode.Type, "Should return document node type")
}

func TestHTTPClient_ParseHTML_InvalidHTML(t *testing.T) {
	client := NewHTTPClient()
	invalidHTML := []byte(`<html><body><unclosed>content`)

	doc, err := client.ParseHTML(invalidHTML)

	require.NoError(t, err, "ParseHTML() should handle invalid HTML gracefully")
	assert.NotNil(t, doc, "ParseHTML() should return document even for invalid HTML")
}

func TestHTTPClient_FetchWebpage_ContentTypeDetection(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		body          string
		shouldSucceed bool
	}{
		{
			name:          "HTML content type",
			contentType:   "text/html",
			body:          "<html><body>Test</body></html>",
			shouldSucceed: true,
		},
		{
			name:          "HTML with charset",
			contentType:   "text/html; charset=utf-8",
			body:          "<html><body>Test</body></html>",
			shouldSucceed: true,
		},
		{
			name:          "No content type",
			contentType:   "",
			body:          "<html><body>Test</body></html>",
			shouldSucceed: true,
		},
		{
			name:          "JSON content type",
			contentType:   "application/json",
			body:          `{"test": "data"}`,
			shouldSucceed: true, // Should still succeed even for non-HTML
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.contentType != "" {
					w.Header().Set("Content-Type", tt.contentType)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewHTTPClient()
			ctx := context.Background()
			content, statusCode, err := client.FetchWebpage(ctx, server.URL)

			if tt.shouldSucceed {
				require.NoError(t, err, "FetchWebpage() should succeed for %s", tt.name)
				assert.NotEmpty(t, content, "FetchWebpage() should return content for %s", tt.name)
				assert.Equal(t, http.StatusOK, statusCode, "Status code should be OK for %s", tt.name)
			} else {
				require.Error(t, err, "FetchWebpage() should fail for %s", tt.name)
			}
		})
	}
}
