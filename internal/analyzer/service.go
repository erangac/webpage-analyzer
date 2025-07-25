package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
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

// service implements the Service interface
type service struct {
	httpClient *http.Client
}

// NewService creates a new instance of the webpage analyzer service
func NewService() Service {
	return &service{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DisableCompression: false,
				DisableKeepAlives:  false,
			},
		},
	}
}

// AnalyzeWebpage analyzes a given webpage using parallel processing
func (s *service) AnalyzeWebpage(ctx context.Context, req AnalysisRequest) (*WebpageAnalysis, error) {
	startTime := time.Now()

	// Create request with proper headers
	httpReq, err := http.NewRequestWithContext(ctx, "GET", req.URL, nil)
	if err != nil {
		return nil, &AnalysisError{
			StatusCode:   0,
			ErrorMessage: fmt.Sprintf("Failed to create request: %v", err),
			URL:          req.URL,
		}
	}

	// Add proper headers
	httpReq.Header.Set("User-Agent", "WebpageAnalyzer/1.0")
	httpReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	httpReq.Header.Set("Accept-Language", "en-US,en;q=0.5")
	httpReq.Header.Set("Accept-Encoding", "gzip, deflate")
	httpReq.Header.Set("Connection", "keep-alive")

	// Fetch the webpage
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, &AnalysisError{
			StatusCode:   0,
			ErrorMessage: fmt.Sprintf("Failed to fetch URL: %v", err),
			URL:          req.URL,
		}
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		return nil, &AnalysisError{
			StatusCode:   resp.StatusCode,
			ErrorMessage: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
			URL:          req.URL,
		}
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &AnalysisError{
			StatusCode:   resp.StatusCode,
			ErrorMessage: fmt.Sprintf("Failed to read response body: %v", err),
			URL:          req.URL,
		}
	}

	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return nil, &AnalysisError{
			StatusCode:   resp.StatusCode,
			ErrorMessage: fmt.Sprintf("Failed to parse HTML: %v", err),
			URL:          req.URL,
		}
	}

	// Initialize analysis result
	analysis := &WebpageAnalysis{
		URL:        req.URL,
		Headings:   make(map[string]int),
		AnalyzedAt: time.Now(),
	}

	// Use parallel processing for different analysis tasks
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Task 1: Extract HTML version (simple, can run independently)
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.extractHTMLVersion(doc, analysis)
	}()

	// Task 2: Extract page title (simple, can run independently)
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.extractPageTitle(doc, analysis)
	}()

	// Task 3: Extract headings (requires thread-safe map access)
	wg.Add(1)
	go func() {
		defer wg.Done()
		headings := make(map[string]int)
		s.extractHeadingsParallel(doc, headings)
		mu.Lock()
		analysis.Headings = headings
		mu.Unlock()
	}()

	// Task 4: Extract links (requires thread-safe counter access)
	wg.Add(1)
	go func() {
		defer wg.Done()
		internal, external, inaccessible := s.extractLinksParallel(doc, req.URL)
		mu.Lock()
		analysis.InternalLinks = internal
		analysis.ExternalLinks = external
		analysis.InaccessibleLinks = inaccessible
		mu.Unlock()
	}()

	// Task 5: Extract login form (simple boolean, can run independently)
	wg.Add(1)
	go func() {
		defer wg.Done()
		hasLogin := s.extractLoginFormParallel(doc)
		mu.Lock()
		analysis.HasLoginForm = hasLogin
		mu.Unlock()
	}()

	// Wait for all analysis tasks to complete
	wg.Wait()

	// Calculate processing time
	analysis.ProcessingTime = time.Since(startTime)

	return analysis, nil
}

// extractHTMLVersion determines the HTML version
func (s *service) extractHTMLVersion(doc *html.Node, analysis *WebpageAnalysis) {
	var findDoctype func(*html.Node)
	findDoctype = func(n *html.Node) {
		if n.Type == html.DoctypeNode {
			if len(n.Attr) > 0 {
				doctype := n.Attr[0].Val
				// Handle different DOCTYPE formats
				if strings.Contains(strings.ToLower(doctype), "html5") || 
				   strings.Contains(strings.ToLower(doctype), "html 5") {
					analysis.HTMLVersion = "HTML5"
				} else if strings.Contains(strings.ToLower(doctype), "html4") || 
				          strings.Contains(strings.ToLower(doctype), "html 4") {
					analysis.HTMLVersion = "HTML4"
				} else if strings.Contains(strings.ToLower(doctype), "xhtml") {
					analysis.HTMLVersion = "XHTML"
				} else {
					analysis.HTMLVersion = doctype
				}
			} else {
				analysis.HTMLVersion = "HTML5 (implied)"
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findDoctype(c)
		}
	}
	findDoctype(doc)
	
	if analysis.HTMLVersion == "" {
		analysis.HTMLVersion = "HTML5 (implied)"
	}
}

// extractPageTitle extracts the page title
func (s *service) extractPageTitle(doc *html.Node, analysis *WebpageAnalysis) {
	var findTitle func(*html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				analysis.PageTitle = strings.TrimSpace(n.FirstChild.Data)
			}
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
		}
	}
	findTitle(doc)
}

// extractHeadingsParallel counts headings by level (thread-safe version)
func (s *service) extractHeadingsParallel(doc *html.Node, headings map[string]int) {
	var countHeadings func(*html.Node)
	countHeadings = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				headings[n.Data]++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			countHeadings(c)
		}
	}
	countHeadings(doc)
}

// extractLinksParallel analyzes internal and external links (thread-safe version)
func (s *service) extractLinksParallel(doc *html.Node, baseURL string) (internal, external, inaccessible int) {
	var analyzeLinks func(*html.Node)
	analyzeLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			hasHref := false
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					hasHref = true
					href := attr.Val
					
					// Skip empty or javascript links
					if href == "" || strings.HasPrefix(href, "javascript:") {
						continue
					}
					
					if strings.HasPrefix(href, "http") {
						external++
					} else if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") {
						internal++
					} else if strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
						// Count as external for now
						external++
					} else {
						// Relative links without leading slash
						internal++
					}
					break
				}
			}
			
			// Check for links without href (potentially inaccessible)
			if !hasHref {
				inaccessible++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			analyzeLinks(c)
		}
	}
	analyzeLinks(doc)
	return internal, external, inaccessible
}

// extractLoginFormParallel checks if the page contains a login form (thread-safe version)
func (s *service) extractLoginFormParallel(doc *html.Node) bool {
	var findLoginForm func(*html.Node) bool
	findLoginForm = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "form" {
			// Check for common login form indicators
			formText := strings.ToLower(s.getNodeText(n))
			if strings.Contains(formText, "login") || strings.Contains(formText, "sign in") || 
			   strings.Contains(formText, "username") || strings.Contains(formText, "password") ||
			   strings.Contains(formText, "email") || strings.Contains(formText, "log in") {
				return true
			}
			
			// Also check for input fields with login-related attributes
			var checkInputs func(*html.Node) bool
			checkInputs = func(node *html.Node) bool {
				if node.Type == html.ElementNode && node.Data == "input" {
					for _, attr := range node.Attr {
						if attr.Key == "type" {
							if attr.Val == "password" {
								return true
							}
						}
						if attr.Key == "name" || attr.Key == "id" {
							name := strings.ToLower(attr.Val)
							if strings.Contains(name, "user") || strings.Contains(name, "pass") ||
							   strings.Contains(name, "login") || strings.Contains(name, "email") {
								return true
							}
						}
					}
				}
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					if checkInputs(c) {
						return true
					}
				}
				return false
			}
			if checkInputs(n) {
				return true
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if findLoginForm(c) {
				return true
			}
		}
		return false
	}
	return findLoginForm(doc)
}

// getNodeText extracts text content from a node
func (s *service) getNodeText(n *html.Node) string {
	var text strings.Builder
	var extractText func(*html.Node)
	extractText = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(n)
	return text.String()
}

// GetAnalysisStatus returns the current status of the analysis service
func (s *service) GetAnalysisStatus(ctx context.Context) (string, error) {
	return "Service is running and ready for parallel webpage analysis", nil
} 