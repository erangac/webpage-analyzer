package analyzer

import (
	"context"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

// Mock HTTP client for testing
type mockHTTPClient struct {
	response string
	error    error
}

func (m *mockHTTPClient) FetchWebpage(ctx context.Context, url string) ([]byte, int, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return nil, 0, ctx.Err()
	default:
		// Continue with normal processing
	}

	if m.error != nil {
		return nil, 500, m.error
	}
	return []byte(m.response), 200, nil
}

func (m *mockHTTPClient) ParseHTML(content []byte) (interface{}, error) {
	doc, err := html.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func TestNewAnalyzerService(t *testing.T) {
	service := NewService()
	if service == nil {
		t.Fatal("NewService() returned nil")
	}
}

func TestAnalyzeWebpage_Success(t *testing.T) {
	// Create a mock HTML response
	htmlContent := `
		<!DOCTYPE html>
		<html>
			<head>
				<title>Test Page</title>
			</head>
			<body>
				<h1>Main Heading</h1>
				<h2>Sub Heading</h2>
				<a href="/internal">Internal Link</a>
				<a href="https://external.com">External Link</a>
				<form>
					<input type="text" name="username">
					<input type="password" name="password">
					<button type="submit">Login</button>
				</form>
			</body>
		</html>
	`

	mockClient := &mockHTTPClient{
		response: htmlContent,
	}

	service := &service{
		httpClient: mockClient,
		htmlParser: NewHTMLParser(),
		workerPool: NewWorkerPool(2),
	}
	defer service.workerPool.Shutdown()

	req := AnalysisRequest{
		URL: "https://example.com",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	if err != nil {
		t.Fatalf("AnalyzeWebpage() returned error: %v", err)
	}

	if result == nil {
		t.Fatal("AnalyzeWebpage() returned nil result")
	}

	// Check HTML version
	if result.HTMLVersion != "HTML5 (implied)" {
		t.Errorf("AnalyzeWebpage() HTMLVersion = %s, want HTML5 (implied)", result.HTMLVersion)
	}

	// Check title
	if result.PageTitle != "Test Page" {
		t.Errorf("AnalyzeWebpage() PageTitle = %s, want 'Test Page'", result.PageTitle)
	}

	// Check headings
	expectedHeadings := map[string]int{"h1": 1, "h2": 1}
	if len(result.Headings) != len(expectedHeadings) {
		t.Errorf("AnalyzeWebpage() Headings count = %d, want %d", len(result.Headings), len(expectedHeadings))
	}
	for heading, count := range expectedHeadings {
		if result.Headings[heading] != count {
			t.Errorf("AnalyzeWebpage() Headings[%s] = %d, want %d", heading, result.Headings[heading], count)
		}
	}

	// Check links
	if result.InternalLinks != 1 {
		t.Errorf("AnalyzeWebpage() InternalLinks = %d, want 1", result.InternalLinks)
	}
	if result.ExternalLinks != 1 {
		t.Errorf("AnalyzeWebpage() ExternalLinks = %d, want 1", result.ExternalLinks)
	}

	// Check login form
	if !result.HasLoginForm {
		t.Error("AnalyzeWebpage() HasLoginForm = false, want true")
	}

	// Check processing time
	if result.ProcessingTime == "" {
		t.Error("AnalyzeWebpage() ProcessingTime is empty")
	}
}

func TestAnalyzeWebpage_HTTPError(t *testing.T) {
	mockClient := &mockHTTPClient{
		error: &AnalysisError{StatusCode: 500, ErrorMessage: "Network error", URL: "https://example.com"},
	}

	service := &service{
		httpClient: mockClient,
		htmlParser: NewHTMLParser(),
		workerPool: NewWorkerPool(2),
	}
	defer service.workerPool.Shutdown()

	req := AnalysisRequest{
		URL: "https://example.com",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	if err == nil {
		t.Fatal("AnalyzeWebpage() should return error")
	}

	if result != nil {
		t.Error("AnalyzeWebpage() should return nil result when error occurs")
	}

	analysisErr, ok := err.(*AnalysisError)
	if !ok {
		t.Fatal("Error should be of type *AnalysisError")
	}

	expectedErrorMsg := "HTTP 500: HTTP 500: Network error (URL: https://example.com) (URL: https://example.com)"
	if analysisErr.Error() != expectedErrorMsg {
		t.Errorf("Error message = %s, want '%s'", analysisErr.Error(), expectedErrorMsg)
	}
}

func TestAnalyzeWebpage_InvalidURL(t *testing.T) {
	service := NewService()

	req := AnalysisRequest{
		URL: "invalid-url",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	if err == nil {
		t.Fatal("AnalyzeWebpage() should return error for invalid URL")
	}

	if result != nil {
		t.Error("AnalyzeWebpage() should return nil result for invalid URL")
	}
}

func TestAnalyzeWebpage_EmptyURL(t *testing.T) {
	service := NewService()

	req := AnalysisRequest{
		URL: "",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	if err == nil {
		t.Fatal("AnalyzeWebpage() should return error for empty URL")
	}

	if result != nil {
		t.Error("AnalyzeWebpage() should return nil result for empty URL")
	}
}

func TestGetAnalysisStatus(t *testing.T) {
	service := NewService()

	ctx := context.Background()
	status, err := service.GetAnalysisStatus(ctx)

	if err != nil {
		t.Fatalf("GetAnalysisStatus() returned error: %v", err)
	}

	if status == "" {
		t.Error("GetAnalysisStatus() returned empty status")
	}

	// Status should be a meaningful string
	if !strings.Contains(strings.ToLower(status), "operational") && 
	   !strings.Contains(strings.ToLower(status), "ready") &&
	   !strings.Contains(strings.ToLower(status), "available") {
		t.Errorf("GetAnalysisStatus() returned unexpected status: %s", status)
	}
}

func TestAnalyzeWebpage_ContextCancellation(t *testing.T) {
	// Create a slow mock client
	mockClient := &mockHTTPClient{
		response: "<html><body>Test</body></html>",
	}

	service := &service{
		httpClient: mockClient,
		htmlParser: NewHTMLParser(),
		workerPool: NewWorkerPool(2),
	}
	defer service.workerPool.Shutdown()

	req := AnalysisRequest{
		URL: "https://example.com",
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	// Cancel the context immediately
	cancel()

	result, err := service.AnalyzeWebpage(ctx, req)

	// Should get context cancellation error
	if err == nil {
		t.Fatal("AnalyzeWebpage() should return error for cancelled context")
	}

	if result != nil {
		t.Error("AnalyzeWebpage() should return nil result for cancelled context")
	}
}

func TestAnalyzeWebpage_ComplexHTML(t *testing.T) {
	// Test with more complex HTML structure
	htmlContent := `
		<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN">
		<html>
			<head>
				<title>Complex Test Page</title>
			</head>
			<body>
				<h1>Main Title</h1>
				<h2>Section 1</h2>
				<h3>Subsection 1.1</h3>
				<h2>Section 2</h2>
				<h3>Subsection 2.1</h3>
				<h4>Sub-subsection</h4>
				
				<nav>
					<a href="/home">Home</a>
					<a href="/about">About</a>
					<a href="https://external-site.com">External</a>
					<a href="mailto:contact@example.com">Contact</a>
					<a href="tel:+1234567890">Phone</a>
					<a href="">Empty Link</a>
					<a href="javascript:void(0)">JavaScript</a>
				</nav>
				
				<main>
					<form id="contact-form">
						<input type="text" name="name" placeholder="Name">
						<input type="email" name="email" placeholder="Email">
						<textarea name="message"></textarea>
						<button type="submit">Send</button>
					</form>
					
					<form id="login-form">
						<h3>Login</h3>
						<input type="text" name="username" placeholder="Username">
						<input type="password" name="password" placeholder="Password">
						<button type="submit">Sign In</button>
					</form>
				</main>
			</body>
		</html>
	`

	mockClient := &mockHTTPClient{
		response: htmlContent,
	}

	service := &service{
		httpClient: mockClient,
		htmlParser: NewHTMLParser(),
		workerPool: NewWorkerPool(2),
	}
	defer service.workerPool.Shutdown()

	req := AnalysisRequest{
		URL: "https://example.com",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	if err != nil {
		t.Fatalf("AnalyzeWebpage() returned error: %v", err)
	}

	// Check HTML version (should be HTML4 based on DOCTYPE)
	if result.HTMLVersion != "HTML4" {
		t.Errorf("AnalyzeWebpage() HTMLVersion = %s, want HTML4", result.HTMLVersion)
	}

	// Check title
	if result.PageTitle != "Complex Test Page" {
		t.Errorf("AnalyzeWebpage() PageTitle = %s, want 'Complex Test Page'", result.PageTitle)
	}

	// Check headings (should have h1, h2, h3, h4)
	expectedHeadings := map[string]int{"h1": 1, "h2": 2, "h3": 3, "h4": 1}
	for heading, count := range expectedHeadings {
		if result.Headings[heading] != count {
			t.Errorf("AnalyzeWebpage() Headings[%s] = %d, want %d", heading, result.Headings[heading], count)
		}
	}

	// Check links
	if result.InternalLinks != 2 { // /home, /about
		t.Errorf("AnalyzeWebpage() InternalLinks = %d, want 2", result.InternalLinks)
	}
	if result.ExternalLinks != 3 { // https://external-site.com, mailto, tel
		t.Errorf("AnalyzeWebpage() ExternalLinks = %d, want 3", result.ExternalLinks)
	}
	if result.InaccessibleLinks != 1 { // empty (javascript is filtered out)
		t.Errorf("AnalyzeWebpage() InaccessibleLinks = %d, want 1", result.InaccessibleLinks)
	}

	// Check login form (should be detected due to password input)
	if !result.HasLoginForm {
		t.Error("AnalyzeWebpage() HasLoginForm = false, want true")
	}
}

func TestAnalyzeWebpage_NoLoginForm(t *testing.T) {
	// Test with HTML that has no login form
	htmlContent := `
		<html>
			<head>
				<title>No Login Page</title>
			</head>
			<body>
				<h1>Welcome</h1>
				<form>
					<input type="text" name="name" placeholder="Name">
					<input type="text" name="comment" placeholder="Comment">
					<button type="submit">Subscribe</button>
				</form>
			</body>
		</html>
	`

	mockClient := &mockHTTPClient{
		response: htmlContent,
	}

	service := &service{
		httpClient: mockClient,
		htmlParser: NewHTMLParser(),
		workerPool: NewWorkerPool(2),
	}
	defer service.workerPool.Shutdown()

	req := AnalysisRequest{
		URL: "https://example.com",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	if err != nil {
		t.Fatalf("AnalyzeWebpage() returned error: %v", err)
	}

	// Should not detect login form
	if result.HasLoginForm {
		t.Error("AnalyzeWebpage() HasLoginForm = true, want false")
	}
} 