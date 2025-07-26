package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient()
	if client == nil {
		t.Fatal("NewHTTPClient() returned nil")
	}
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

	if err != nil {
		t.Fatalf("FetchWebpage() returned error: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Errorf("FetchWebpage() status code = %d, want %d", statusCode, http.StatusOK)
	}

	if len(content) == 0 {
		t.Fatal("FetchWebpage() returned empty content")
	}

	// Check if the HTML contains expected content
	htmlStr := string(content)
	if !strings.Contains(htmlStr, "Test Page") {
		t.Errorf("FetchWebpage() response does not contain expected title: %s", htmlStr)
	}

	if !strings.Contains(htmlStr, "Hello World") {
		t.Errorf("FetchWebpage() response does not contain expected body: %s", htmlStr)
	}
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

	if err != nil {
		t.Fatalf("FetchWebpage() returned error: %v", err)
	}

	if statusCode != http.StatusNotFound {
		t.Errorf("FetchWebpage() status code = %d, want %d", statusCode, http.StatusNotFound)
	}

	if len(content) == 0 {
		t.Error("FetchWebpage() should return content even for 404")
	}
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

	if err != nil {
		t.Fatalf("FetchWebpage() returned error: %v", err)
	}

	if statusCode != http.StatusInternalServerError {
		t.Errorf("FetchWebpage() status code = %d, want %d", statusCode, http.StatusInternalServerError)
	}

	if len(content) == 0 {
		t.Error("FetchWebpage() should return content even for 500")
	}
}

func TestHTTPClient_FetchWebpage_InvalidURL(t *testing.T) {
	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, "invalid-url")

	if err == nil {
		t.Fatal("FetchWebpage() should return error for invalid URL")
	}

	if content != nil {
		t.Error("FetchWebpage() should return nil content for invalid URL")
	}

	if statusCode != 400 {
		t.Errorf("FetchWebpage() status code = %d, want 400", statusCode)
	}
}

func TestHTTPClient_FetchWebpage_EmptyURL(t *testing.T) {
	client := NewHTTPClient()
	ctx := context.Background()
	content, statusCode, err := client.FetchWebpage(ctx, "")

	if err == nil {
		t.Fatal("FetchWebpage() should return error for empty URL")
	}

	if content != nil {
		t.Error("FetchWebpage() should return nil content for empty URL")
	}

	if statusCode != 400 {
		t.Errorf("FetchWebpage() status code = %d, want 400", statusCode)
	}
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

	if err == nil {
		t.Fatal("FetchWebpage() should return error for timeout")
	}

	if content != nil {
		t.Error("FetchWebpage() should return nil content for timeout")
	}

	if statusCode != 408 {
		t.Errorf("FetchWebpage() status code = %d, want 408", statusCode)
	}
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

	if err != nil {
		t.Fatalf("FetchWebpage() should not return error for non-HTML content: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Errorf("FetchWebpage() status code = %d, want %d", statusCode, http.StatusOK)
	}

	if len(content) == 0 {
		t.Fatal("FetchWebpage() should return content even if it's not HTML")
	}

	// Should still return the content even if it's not HTML
	contentStr := string(content)
	if !strings.Contains(contentStr, "This is JSON, not HTML") {
		t.Errorf("FetchWebpage() should return the content: %s", contentStr)
	}
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

	if err != nil {
		t.Fatalf("FetchWebpage() returned error for large response: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Errorf("FetchWebpage() status code = %d, want %d", statusCode, http.StatusOK)
	}

	if len(content) == 0 {
		t.Fatal("FetchWebpage() returned empty content for large response")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Large Page") {
		t.Error("FetchWebpage() should contain the title from large response")
	}

	if len(contentStr) < 1000 {
		t.Error("FetchWebpage() should return the full large response")
	}
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

	if err != nil {
		t.Fatalf("FetchWebpage() returned error for redirect: %v", err)
	}

	if statusCode != http.StatusOK {
		t.Errorf("FetchWebpage() status code = %d, want %d", statusCode, http.StatusOK)
	}

	if len(content) == 0 {
		t.Fatal("FetchWebpage() returned empty content for redirect")
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Final Page") {
		t.Error("FetchWebpage() should follow redirect and return final page content")
	}
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

	if err != nil {
		t.Fatalf("FetchWebpage() returned error: %v", err)
	}

	if userAgent == "" {
		t.Error("FetchWebpage() should set User-Agent header")
	}

	if !strings.Contains(userAgent, "WebpageAnalyzer") {
		t.Errorf("User-Agent should contain 'WebpageAnalyzer', got: %s", userAgent)
	}
}

func TestHTTPClient_ParseHTML_Success(t *testing.T) {
	client := NewHTTPClient()
	htmlContent := []byte(`<!DOCTYPE html><html><head><title>Test</title></head><body>Hello</body></html>`)

	doc, err := client.ParseHTML(htmlContent)

	if err != nil {
		t.Fatalf("ParseHTML() returned error: %v", err)
	}

	if doc == nil {
		t.Fatal("ParseHTML() returned nil document")
	}

	// Check if it's a valid HTML node
	htmlNode, ok := doc.(*html.Node)
	if !ok {
		t.Fatal("ParseHTML() should return *html.Node")
	}

	if htmlNode.Type != html.DocumentNode {
		t.Errorf("ParseHTML() returned node type %d, want %d", htmlNode.Type, html.DocumentNode)
	}
}

func TestHTTPClient_ParseHTML_InvalidHTML(t *testing.T) {
	client := NewHTTPClient()
	invalidHTML := []byte(`<html><body><unclosed>content`)

	doc, err := client.ParseHTML(invalidHTML)

	if err != nil {
		t.Fatalf("ParseHTML() should handle invalid HTML gracefully: %v", err)
	}

	if doc == nil {
		t.Fatal("ParseHTML() should return document even for invalid HTML")
	}
}

func TestHTTPClient_FetchWebpage_ContentTypeDetection(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		body           string
		shouldSucceed  bool
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

			if tt.shouldSucceed && err != nil {
				t.Errorf("FetchWebpage() should succeed for %s, got error: %v", tt.name, err)
			}

			if !tt.shouldSucceed && err == nil {
				t.Errorf("FetchWebpage() should fail for %s", tt.name)
			}

			if tt.shouldSucceed && len(content) == 0 {
				t.Errorf("FetchWebpage() should return content for %s", tt.name)
			}

			if tt.shouldSucceed && statusCode != http.StatusOK {
				t.Errorf("FetchWebpage() status code = %d, want %d for %s", statusCode, http.StatusOK, tt.name)
			}
		})
	}
} 