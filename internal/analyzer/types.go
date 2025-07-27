package analyzer

import (
	"context"
	"fmt"
	"time"
)

// WebpageAnalysis represents the result of analyzing a webpage.
// @Description Comprehensive result of webpage analysis
type WebpageAnalysis struct {
	URL               string         `json:"url" example:"https://example.com"`
	HTMLVersion       string         `json:"html_version" example:"HTML5"`
	PageTitle         string         `json:"page_title" example:"Example Domain"`
	Headings          map[string]int `json:"headings"` // level -> count.
	InternalLinks     int            `json:"internal_links" example:"15"`
	ExternalLinks     int            `json:"external_links" example:"8"`
	InaccessibleLinks int            `json:"inaccessible_links" example:"0"`
	HasLoginForm      bool           `json:"has_login_form" example:"false"`
	AnalyzedAt        time.Time      `json:"analyzed_at" example:"2024-01-15T10:30:00Z"`
	ProcessingTime    string         `json:"processing_time" example:"150ms"`
}

// AnalysisRequest represents a request to analyze a webpage.
// @Description Request to analyze a webpage
type AnalysisRequest struct {
	URL string `json:"url" example:"https://example.com" binding:"required"`
}

// AnalysisError represents an error during webpage analysis.
// @Description Detailed error response when webpage analysis fails
type AnalysisError struct {
	StatusCode   int    `json:"status_code" example:"404"`
	ErrorMessage string `json:"error_message" example:"Not Found: The requested webpage could not be found on the server."`
	URL          string `json:"url" example:"https://nonexistent.example.com"`
}

// Error implements the error interface.
func (e *AnalysisError) Error() string {
	return fmt.Sprintf("HTTP %d: %s (URL: %s)", e.StatusCode, e.ErrorMessage, e.URL)
}

// Service defines the interface for webpage analysis operations.
type Service interface {
	AnalyzeWebpage(ctx context.Context, req AnalysisRequest) (*WebpageAnalysis, error)
	GetAnalysisStatus(ctx context.Context) (string, error)
}
