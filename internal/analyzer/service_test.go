package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"webpage-analyzer/internal/parser"
	"webpage-analyzer/internal/worker"
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
	// Create mock dependencies
	mockClient := &mockHTTPClient{
		response: `
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
				</body>
			</html>
		`,
	}

	htmlParser := parser.NewHTMLParser()
	workerPool := worker.NewWorkerPool(2)

	// Create service with mock dependencies
	service := NewServiceWithDependencies(mockClient, htmlParser, workerPool)

	// Test analysis
	ctx := context.Background()
	req := AnalysisRequest{URL: "https://example.com"}
	result, err := service.AnalyzeWebpage(ctx, req)

	require.NoError(t, err, "AnalyzeWebpage() should not return error")
	require.NotNil(t, result, "AnalyzeWebpage() should not return nil result")
	assert.Equal(t, "HTML5 (implied)", result.HTMLVersion, "HTML version should match")
	assert.Equal(t, "Test Page", result.PageTitle, "Page title should match")
	assert.Equal(t, 1, result.Headings["h1"], "H1 count should match")
	assert.Equal(t, 1, result.Headings["h2"], "H2 count should match")
	assert.Equal(t, 1, result.InternalLinks, "Internal links count should match")
	assert.Equal(t, 1, result.ExternalLinks, "External links count should match")
	assert.False(t, result.HasLoginForm, "Login form should not be detected")
}

func TestAnalyzeWebpage_HTTPError(t *testing.T) {
	// Create mock client that returns error
	mockClient := &mockHTTPClient{
		error: assert.AnError,
	}

	htmlParser := parser.NewHTMLParser()
	workerPool := worker.NewWorkerPool(2)

	service := NewServiceWithDependencies(mockClient, htmlParser, workerPool)

	ctx := context.Background()
	req := AnalysisRequest{URL: "https://example.com"}
	result, err := service.AnalyzeWebpage(ctx, req)

	require.Error(t, err, "AnalyzeWebpage() should return error")
	assert.Nil(t, result, "AnalyzeWebpage() should return nil result on error")

	// Check if it's an AnalysisError
	analysisErr, ok := err.(*AnalysisError)
	require.True(t, ok, "Error should be of type AnalysisError")
	assert.Equal(t, 500, analysisErr.StatusCode, "Status code should match")
}

func TestAnalyzeWebpage_InvalidURL(t *testing.T) {
	// Create mock client that returns error for invalid URL
	mockClient := &mockHTTPClient{
		error: assert.AnError,
	}

	htmlParser := parser.NewHTMLParser()
	workerPool := worker.NewWorkerPool(2)

	service := NewServiceWithDependencies(mockClient, htmlParser, workerPool)

	ctx := context.Background()
	req := AnalysisRequest{URL: "invalid-url"}
	result, err := service.AnalyzeWebpage(ctx, req)

	require.Error(t, err, "AnalyzeWebpage() should return error for invalid URL")
	assert.Nil(t, result, "AnalyzeWebpage() should return nil result for invalid URL")
}

func TestAnalyzeWebpage_EmptyURL(t *testing.T) {
	// Create mock client that returns error for empty URL
	mockClient := &mockHTTPClient{
		error: assert.AnError,
	}

	htmlParser := parser.NewHTMLParser()
	workerPool := worker.NewWorkerPool(2)

	service := NewServiceWithDependencies(mockClient, htmlParser, workerPool)

	ctx := context.Background()
	req := AnalysisRequest{URL: ""}
	result, err := service.AnalyzeWebpage(ctx, req)

	require.Error(t, err, "AnalyzeWebpage() should return error for empty URL")
	assert.Nil(t, result, "AnalyzeWebpage() should return nil result for empty URL")
}

func TestGetAnalysisStatus(t *testing.T) {
	service := NewService()
	ctx := context.Background()

	status, err := service.GetAnalysisStatus(ctx)

	require.NoError(t, err, "GetAnalysisStatus() should not return error")
	assert.NotEmpty(t, status, "GetAnalysisStatus() should return non-empty status")
	assert.Contains(t, status, "Service is running", "Status should contain expected message")
}

func TestAnalyzeWebpage_ComplexHTML(t *testing.T) {
	// Create mock client with complex HTML
	mockClient := &mockHTTPClient{
		response: `
			<!DOCTYPE html>
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
					<h3>Subsection 2.2</h3>
					
					<a href="/page1">Internal Page 1</a>
					<a href="/page2">Internal Page 2</a>
					<a href="https://google.com">Google</a>
					<a href="https://github.com">GitHub</a>
					<a href="mailto:test@example.com">Email</a>
					<a href="tel:+1234567890">Phone</a>
					<a href="">Empty Link</a>
					<a href="javascript:void(0)">JavaScript</a>
					
					<form>
						<input type="text" name="username" placeholder="Username">
						<input type="password" name="password" placeholder="Password">
						<button type="submit">Login</button>
					</form>
				</body>
			</html>
		`,
	}

	htmlParser := parser.NewHTMLParser()
	workerPool := worker.NewWorkerPool(2)

	service := NewServiceWithDependencies(mockClient, htmlParser, workerPool)

	ctx := context.Background()
	req := AnalysisRequest{URL: "https://example.com"}
	result, err := service.AnalyzeWebpage(ctx, req)

	require.NoError(t, err, "AnalyzeWebpage() should not return error")
	require.NotNil(t, result, "AnalyzeWebpage() should not return nil result")

	// Check complex analysis results
	assert.Equal(t, "HTML5 (implied)", result.HTMLVersion, "HTML version should match")
	assert.Equal(t, "Complex Test Page", result.PageTitle, "Page title should match")

	// Check headings
	assert.Equal(t, 1, result.Headings["h1"], "H1 count should match")
	assert.Equal(t, 2, result.Headings["h2"], "H2 count should match")
	assert.Equal(t, 3, result.Headings["h3"], "H3 count should match")

	// Check links (updated for new link categorization logic)
	assert.Equal(t, 4, result.InternalLinks, "Internal links count should match")
	assert.Equal(t, 2, result.ExternalLinks, "External links count should match")
	assert.Equal(t, 2, result.InaccessibleLinks, "Inaccessible links count should match")

	// Check login form detection
	assert.True(t, result.HasLoginForm, "Login form should be detected")
}

func TestAnalyzeWebpage_NoLoginForm(t *testing.T) {
	// Create mock client with HTML without login form
	mockClient := &mockHTTPClient{
		response: `
			<!DOCTYPE html>
			<html>
				<head>
					<title>No Login Form</title>
				</head>
				<body>
					<h1>Welcome</h1>
					<p>This page has no login form.</p>
					<a href="/about">About</a>
				</body>
			</html>
		`,
	}

	htmlParser := parser.NewHTMLParser()
	workerPool := worker.NewWorkerPool(2)

	service := NewServiceWithDependencies(mockClient, htmlParser, workerPool)

	ctx := context.Background()
	req := AnalysisRequest{URL: "https://example.com"}
	result, err := service.AnalyzeWebpage(ctx, req)

	require.NoError(t, err, "AnalyzeWebpage() should not return error")
	require.NotNil(t, result, "AnalyzeWebpage() should not return nil result")
	assert.False(t, result.HasLoginForm, "Login form should not be detected")
}
