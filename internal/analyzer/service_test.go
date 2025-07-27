package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NotNil(t, service, "NewService() should not return nil")
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

	require.NoError(t, err, "AnalyzeWebpage() should not return error")
	require.NotNil(t, result, "AnalyzeWebpage() should not return nil result")

	// Check HTML version
	assert.Equal(t, "HTML5 (implied)", result.HTMLVersion, "HTML version should match")

	// Check title
	assert.Equal(t, "Test Page", result.PageTitle, "Page title should match")

	// Check headings
	expectedHeadings := map[string]int{"h1": 1, "h2": 1}
	assert.Len(t, result.Headings, len(expectedHeadings), "Headings count should match")
	for heading, count := range expectedHeadings {
		assert.Equal(t, count, result.Headings[heading], "Heading count for %s should match", heading)
	}

	// Check links
	assert.Equal(t, 1, result.InternalLinks, "Internal links count should match")
	assert.Equal(t, 1, result.ExternalLinks, "External links count should match")

	// Check login form
	assert.True(t, result.HasLoginForm, "Login form should be detected")

	// Check processing time
	assert.NotEmpty(t, result.ProcessingTime, "Processing time should not be empty")
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

	require.Error(t, err, "AnalyzeWebpage() should return error")
	assert.Nil(t, result, "AnalyzeWebpage() should return nil result when error occurs")

	analysisErr, ok := err.(*AnalysisError)
	require.True(t, ok, "Error should be of type *AnalysisError")

	expectedErrorMsg := "HTTP 500: HTTP 500: Network error (URL: https://example.com) (URL: https://example.com)"
	assert.Equal(t, expectedErrorMsg, analysisErr.Error(), "Error message should match expected")
}

func TestAnalyzeWebpage_InvalidURL(t *testing.T) {
	service := NewService()

	req := AnalysisRequest{
		URL: "invalid-url",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	require.Error(t, err, "AnalyzeWebpage() should return error for invalid URL")
	assert.Nil(t, result, "AnalyzeWebpage() should return nil result for invalid URL")
}

func TestAnalyzeWebpage_EmptyURL(t *testing.T) {
	service := NewService()

	req := AnalysisRequest{
		URL: "",
	}

	ctx := context.Background()
	result, err := service.AnalyzeWebpage(ctx, req)

	require.Error(t, err, "AnalyzeWebpage() should return error for empty URL")
	assert.Nil(t, result, "AnalyzeWebpage() should return nil result for empty URL")
}

func TestGetAnalysisStatus(t *testing.T) {
	service := NewService()

	ctx := context.Background()
	status, err := service.GetAnalysisStatus(ctx)

	require.NoError(t, err, "GetAnalysisStatus() should not return error")
	assert.NotEmpty(t, status, "GetAnalysisStatus() should not return empty status")

	// Status should be a meaningful string
	statusLower := strings.ToLower(status)
	assert.True(t,
		strings.Contains(statusLower, "operational") ||
			strings.Contains(statusLower, "ready") ||
			strings.Contains(statusLower, "available"),
		"GetAnalysisStatus() should return meaningful status, got: %s", status)
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

	require.NoError(t, err, "AnalyzeWebpage() should not return error")
	require.NotNil(t, result, "AnalyzeWebpage() should not return nil result")

	// Check HTML version (should be HTML4 based on DOCTYPE)
	assert.Equal(t, "HTML4", result.HTMLVersion, "HTML version should be HTML4")

	// Check title
	assert.Equal(t, "Complex Test Page", result.PageTitle, "Page title should match")

	// Check headings (should have h1, h2, h3, h4)
	expectedHeadings := map[string]int{"h1": 1, "h2": 2, "h3": 3, "h4": 1}
	for heading, count := range expectedHeadings {
		assert.Equal(t, count, result.Headings[heading], "Heading count for %s should match", heading)
	}

	// Check links
	assert.Equal(t, 2, result.InternalLinks, "Internal links count should match")         // /home, /about
	assert.Equal(t, 3, result.ExternalLinks, "External links count should match")         // https://external-site.com, mailto, tel
	assert.Equal(t, 1, result.InaccessibleLinks, "Inaccessible links count should match") // empty (javascript is filtered out)

	// Check login form (should be detected due to password input)
	assert.True(t, result.HasLoginForm, "Login form should be detected")
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

	require.NoError(t, err, "AnalyzeWebpage() should not return error")
	require.NotNil(t, result, "AnalyzeWebpage() should not return nil result")

	// Should not detect login form
	assert.False(t, result.HasLoginForm, "Login form should not be detected")
}
