package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// service implements the Service interface
type service struct {
	httpClient    HTTPClient
	htmlParser    HTMLParser
	workerPool    *WorkerPool
}

// NewService creates a new instance of the webpage analyzer service
func NewService() Service {
	return &service{
		httpClient: NewHTTPClient(),
		htmlParser: NewHTMLParser(),
		workerPool: NewWorkerPool(5), // 5 workers for analysis tasks
	}
}

// NewServiceWithDependencies creates a service with custom dependencies (useful for testing)
func NewServiceWithDependencies(httpClient HTTPClient, htmlParser HTMLParser, workerPool *WorkerPool) Service {
	return &service{
		httpClient: httpClient,
		htmlParser: htmlParser,
		workerPool: workerPool,
	}
}

// AnalyzeWebpage analyzes a given webpage using the worker pool
func (s *service) AnalyzeWebpage(ctx context.Context, req AnalysisRequest) (*WebpageAnalysis, error) {
	startTime := time.Now()

	// Fetch the webpage
	body, statusCode, err := s.httpClient.FetchWebpage(ctx, req.URL)
	if err != nil {
		// Create a more meaningful error response
		return nil, &AnalysisError{
			StatusCode:   statusCode,
			ErrorMessage: err.Error(),
			URL:          req.URL,
		}
	}

	// Check if the response is successful
	if statusCode != http.StatusOK {
		// Provide specific error messages for different HTTP status codes
		errorMessage := s.getHTTPStatusMessage(statusCode)
		return nil, &AnalysisError{
			StatusCode:   statusCode,
			ErrorMessage: errorMessage,
			URL:          req.URL,
		}
	}

	// Parse the HTML
	doc, err := s.httpClient.ParseHTML(body)
	if err != nil {
		return nil, &AnalysisError{
			StatusCode:   statusCode,
			ErrorMessage: fmt.Sprintf("Failed to parse HTML content: %v", err),
			URL:          req.URL,
		}
	}

	// Initialize analysis result
	analysis := &WebpageAnalysis{
		URL:        req.URL,
		Headings:   make(map[string]int),
		AnalyzedAt: time.Now(),
	}

	// Use worker pool for parallel analysis
	taskGroup := NewAnalysisTaskGroup(s.workerPool)

	// Add analysis tasks to the group
	taskGroup.AddTask("html_version", func() (interface{}, error) {
		version := s.htmlParser.ExtractHTMLVersion(doc)
		return version, nil
	})

	taskGroup.AddTask("page_title", func() (interface{}, error) {
		title := s.htmlParser.ExtractPageTitle(doc)
		return title, nil
	})

	taskGroup.AddTask("headings", func() (interface{}, error) {
		headings := s.htmlParser.ExtractHeadings(doc)
		return headings, nil
	})

	taskGroup.AddTask("links", func() (interface{}, error) {
		internal, external, inaccessible := s.htmlParser.ExtractLinks(doc, req.URL)
		return map[string]int{
			"internal":     internal,
			"external":     external,
			"inaccessible": inaccessible,
		}, nil
	})

	taskGroup.AddTask("login_form", func() (interface{}, error) {
		hasLogin := s.htmlParser.ExtractLoginForm(doc)
		return hasLogin, nil
	})

	// Execute all tasks in parallel
	taskGroup.ExecuteAll()

	// Collect results
	if htmlVersion, err := taskGroup.GetResult("html_version"); err == nil {
		analysis.HTMLVersion = htmlVersion.(string)
	}

	if pageTitle, err := taskGroup.GetResult("page_title"); err == nil {
		analysis.PageTitle = pageTitle.(string)
	}

	if headings, err := taskGroup.GetResult("headings"); err == nil {
		analysis.Headings = headings.(map[string]int)
	}

	if links, err := taskGroup.GetResult("links"); err == nil {
		linkMap := links.(map[string]int)
		analysis.InternalLinks = linkMap["internal"]
		analysis.ExternalLinks = linkMap["external"]
		analysis.InaccessibleLinks = linkMap["inaccessible"]
	}

	if hasLogin, err := taskGroup.GetResult("login_form"); err == nil {
		analysis.HasLoginForm = hasLogin.(bool)
	}

	// Calculate processing time
	analysis.ProcessingTime = time.Since(startTime)

	return analysis, nil
}

// getHTTPStatusMessage returns a user-friendly message for HTTP status codes
func (s *service) getHTTPStatusMessage(statusCode int) string {
	switch statusCode {
	case 400:
		return "Bad Request: The URL format is invalid or the request is malformed."
	case 401:
		return "Unauthorized: Access to this resource requires authentication."
	case 403:
		return "Forbidden: Access to this resource is denied."
	case 404:
		return "Not Found: The requested webpage could not be found on the server."
	case 408:
		return "Request Timeout: The server took too long to respond."
	case 429:
		return "Too Many Requests: The server is receiving too many requests. Please try again later."
	case 495:
		return "SSL Certificate Error: There was a problem with the security certificate."
	case 500:
		return "Internal Server Error: The server encountered an error while processing the request."
	case 502:
		return "Bad Gateway: The server received an invalid response from an upstream server."
	case 503:
		return "Service Unavailable: The server is temporarily unable to handle the request."
	case 504:
		return "Gateway Timeout: The server acting as a gateway did not receive a timely response."
	default:
		return fmt.Sprintf("HTTP %d: %s", statusCode, http.StatusText(statusCode))
	}
}

// GetAnalysisStatus returns the current status of the analysis service
func (s *service) GetAnalysisStatus(ctx context.Context) (string, error) {
	return "Service is running and ready for parallel webpage analysis with worker pool", nil
} 