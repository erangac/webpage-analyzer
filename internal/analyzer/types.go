package analyzer

import (
	"context"
	"fmt"
	"time"
)

// WebpageAnalysis represents the result of analyzing a webpage
type WebpageAnalysis struct {
	URL              string            `json:"url"`
	HTMLVersion      string            `json:"html_version"`
	PageTitle        string            `json:"page_title"`
	Headings         map[string]int    `json:"headings"` // level -> count
	InternalLinks    int               `json:"internal_links"`
	ExternalLinks    int               `json:"external_links"`
	InaccessibleLinks int              `json:"inaccessible_links"`
	HasLoginForm     bool              `json:"has_login_form"`
	AnalyzedAt       time.Time         `json:"analyzed_at"`
	ProcessingTime   time.Duration     `json:"processing_time"`
}

// AnalysisRequest represents a request to analyze a webpage
type AnalysisRequest struct {
	URL string `json:"url"`
}

// AnalysisError represents an error during webpage analysis
type AnalysisError struct {
	StatusCode    int    `json:"status_code"`
	ErrorMessage  string `json:"error_message"`
	URL           string `json:"url"`
}

// Error implements the error interface
func (e *AnalysisError) Error() string {
	return fmt.Sprintf("HTTP %d: %s (URL: %s)", e.StatusCode, e.ErrorMessage, e.URL)
}

// Service defines the interface for webpage analysis operations
type Service interface {
	AnalyzeWebpage(ctx context.Context, req AnalysisRequest) (*WebpageAnalysis, error)
	GetAnalysisStatus(ctx context.Context) (string, error)
}

// HTMLParser defines the interface for HTML parsing operations
type HTMLParser interface {
	ExtractHTMLVersion(doc interface{}) string
	ExtractPageTitle(doc interface{}) string
	ExtractHeadings(doc interface{}) map[string]int
	ExtractLinks(doc interface{}, baseURL string) (internal, external, inaccessible int)
	ExtractLoginForm(doc interface{}) bool
}

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	FetchWebpage(ctx context.Context, url string) ([]byte, int, error)
	ParseHTML(content []byte) (interface{}, error)
}

// WorkerPoolManager defines the interface for worker pool operations
type WorkerPoolManager interface {
	Submit(task Task)
	SubmitAndWait(task Task) error
	Wait()
	Shutdown()
} 