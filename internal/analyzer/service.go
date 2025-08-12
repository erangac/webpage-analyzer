package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"webpage-analyzer/internal/cache"
	"webpage-analyzer/internal/client"
	"webpage-analyzer/internal/parser"
	"webpage-analyzer/internal/worker"
)

// service implements the Service interface.
type service struct {
	httpClient client.HTTPClient
	htmlParser parser.HTMLParser
	workerPool *worker.WorkerPool
	cache      *cache.LocalCache[string, *WebpageAnalysis]
}

// NewService creates a new instance of the webpage analyzer service.
// allow to pass in a cache interface
func NewService(cache *cache.LocalCache[string, *WebpageAnalysis]) Service {
	return &service{
		httpClient: client.NewHTTPClient(),
		htmlParser: parser.NewHTMLParser(),
		workerPool: worker.NewWorkerPool(5), // 5 workers for analysis tasks.
		cache:      cache,
	}
}

// NewServiceWithDependencies creates a service with custom dependencies (useful for testing).
func NewServiceWithDependencies(httpClient client.HTTPClient, htmlParser parser.HTMLParser, workerPool *worker.WorkerPool) Service {
	return &service{
		httpClient: httpClient,
		htmlParser: htmlParser,
		workerPool: workerPool,
	}
}

// AnalyzeWebpage analyzes a given webpage using the worker pool.
func (s *service) AnalyzeWebpage(ctx context.Context, req AnalysisRequest) (*WebpageAnalysis, error) {
	startTime := time.Now()
	slog.Info("Starting webpage analysis", "url", req.URL)

	// Check if the url is already in the cache.
	if cached, ok := s.cache.Get(req.URL); ok {
		slog.Info("Returning cached analysis", "url", req.URL)
		return cached, nil
	}

	// Fetch the webpage.
	slog.Info("Fetching webpage content", "url", req.URL)
	body, statusCode, err := s.httpClient.FetchWebpage(ctx, req.URL)
	if err != nil {
		slog.Error("Error fetching webpage", "url", req.URL, "error", err, "status_code", statusCode)
		// Create a more meaningful error response.
		return nil, &AnalysisError{
			StatusCode:   statusCode,
			ErrorMessage: err.Error(),
			URL:          req.URL,
		}
	}
	slog.Info("Successfully fetched webpage", "url", req.URL, "status_code", statusCode, "body_size_bytes", len(body))

	// Check if the response is successful.
	if statusCode != http.StatusOK {
		slog.Error("HTTP error", "url", req.URL, "status_code", statusCode)
		// Provide specific error messages for different HTTP status codes.
		errorMessage := s.getHTTPStatusMessage(statusCode)
		return nil, &AnalysisError{
			StatusCode:   statusCode,
			ErrorMessage: errorMessage,
			URL:          req.URL,
		}
	}

	// Parse the HTML.
	slog.Info("Parsing HTML content", "url", req.URL)
	doc, err := s.httpClient.ParseHTML(body)
	if err != nil {
		slog.Error("Error parsing HTML", "url", req.URL, "error", err)
		return nil, &AnalysisError{
			StatusCode:   statusCode,
			ErrorMessage: fmt.Sprintf("Failed to parse HTML content: %v", err),
			URL:          req.URL,
		}
	}
	slog.Info("Successfully parsed HTML", "url", req.URL)

	// Initialize analysis result.
	analysis := &WebpageAnalysis{
		URL:        req.URL,
		Headings:   make(map[string]int),
		AnalyzedAt: time.Now(),
	}

	// Use worker pool for parallel analysis.
	slog.Info("Starting parallel analysis tasks", "url", req.URL)
	taskGroup := worker.NewAnalysisTaskGroup(s.workerPool)

	// Add analysis tasks to the group.
	taskGroup.AddTask("html_version", func() (interface{}, error) {
		slog.Info("Extracting HTML version", "url", req.URL)
		version := s.htmlParser.ExtractHTMLVersion(doc)
		slog.Info("HTML version extracted", "url", req.URL, "version", version)
		return version, nil
	})

	taskGroup.AddTask("page_title", func() (interface{}, error) {
		slog.Info("Extracting page title", "url", req.URL)
		title := s.htmlParser.ExtractPageTitle(doc)
		slog.Info("Page title extracted", "url", req.URL, "title", title)
		return title, nil
	})

	taskGroup.AddTask("headings", func() (interface{}, error) {
		slog.Info("Extracting headings", "url", req.URL)
		headings := s.htmlParser.ExtractHeadings(doc)
		slog.Info("Headings extracted", "url", req.URL, "heading_types_count", len(headings))
		return headings, nil
	})

	taskGroup.AddTask("links", func() (interface{}, error) {
		slog.Info("Extracting links", "url", req.URL)
		internal, external, inaccessible := s.htmlParser.ExtractLinks(doc, req.URL)
		slog.Info("Links extracted", "url", req.URL, "internal_count", internal, "external_count", external, "inaccessible_count", inaccessible)
		return map[string]int{
			"internal":     internal,
			"external":     external,
			"inaccessible": inaccessible,
		}, nil
	})

	taskGroup.AddTask("login_form", func() (interface{}, error) {
		slog.Info("Checking for login form", "url", req.URL)
		hasLogin := s.htmlParser.ExtractLoginForm(doc)
		slog.Info("Login form check completed", "url", req.URL, "has_login_form", hasLogin)
		return hasLogin, nil
	})

	// Execute all tasks in parallel.
	slog.Info("Executing analysis tasks in parallel", "url", req.URL, "task_count", 5)
	taskGroup.ExecuteAll()
	slog.Info("All analysis tasks completed", "url", req.URL)

	// Collect results.
	slog.Info("Collecting analysis results", "url", req.URL)

	if htmlVersion, err := taskGroup.GetResult("html_version"); err == nil {
		analysis.HTMLVersion = htmlVersion.(string)
		slog.Info("HTML version result collected", "url", req.URL, "version", analysis.HTMLVersion)
	} else {
		slog.Error("Error getting HTML version result", "url", req.URL, "error", err)
	}

	if pageTitle, err := taskGroup.GetResult("page_title"); err == nil {
		analysis.PageTitle = pageTitle.(string)
		slog.Info("Page title result collected", "url", req.URL, "title", analysis.PageTitle)
	} else {
		slog.Error("Error getting page title result", "url", req.URL, "error", err)
	}

	if headings, err := taskGroup.GetResult("headings"); err == nil {
		analysis.Headings = headings.(map[string]int)
		slog.Info("Headings result collected", "url", req.URL, "headings", analysis.Headings)
	} else {
		slog.Error("Error getting headings result", "url", req.URL, "error", err)
	}

	if links, err := taskGroup.GetResult("links"); err == nil {
		linkMap := links.(map[string]int)
		analysis.InternalLinks = linkMap["internal"]
		analysis.ExternalLinks = linkMap["external"]
		analysis.InaccessibleLinks = linkMap["inaccessible"]
		slog.Info("Links result collected", "url", req.URL, "internal_count", analysis.InternalLinks, "external_count", analysis.ExternalLinks, "inaccessible_count", analysis.InaccessibleLinks)
	} else {
		slog.Error("Error getting links result", "url", req.URL, "error", err)
	}

	if hasLogin, err := taskGroup.GetResult("login_form"); err == nil {
		analysis.HasLoginForm = hasLogin.(bool)
		slog.Info("Login form result collected", "url", req.URL, "has_login_form", analysis.HasLoginForm)
	} else {
		slog.Error("Error getting login form result", "url", req.URL, "error", err)
	}

	// Calculate processing time.
	analysis.ProcessingTime = time.Since(startTime).String()
	slog.Info("Analysis completed", "url", req.URL, "processing_time", analysis.ProcessingTime)

	// Save the analysis to the cache.
	slog.Info("Saving analysis to cache", "url", req.URL, "analysis", analysis)
	s.cache.Set(req.URL, analysis)

	return analysis, nil
}

// getHTTPStatusMessage returns a user-friendly message for HTTP status codes.
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

// GetAnalysisStatus returns the current status of the analysis service.
func (s *service) GetAnalysisStatus(ctx context.Context) (string, error) {
	slog.Info("Service status requested")
	status := "Service is running and ready for parallel webpage analysis with worker pool"
	slog.Info("Service status", "status", status)
	return status, nil
}
