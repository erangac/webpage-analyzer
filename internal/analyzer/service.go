package analyzer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// AnalyzeWebpage analyzes a given webpage and returns detailed information
func (s *service) AnalyzeWebpage(ctx context.Context, req AnalysisRequest) (*WebpageAnalysis, error) {
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

	// Analyze the webpage
	analysis := &WebpageAnalysis{
		URL:        req.URL,
		Headings:   make(map[string]int),
		AnalyzedAt: time.Now(),
	}

	// Extract information
	s.extractHTMLVersion(doc, analysis)
	s.extractPageTitle(doc, analysis)
	s.extractHeadings(doc, analysis)
	s.extractLinks(doc, analysis, req.URL)
	s.extractLoginForm(doc, analysis)

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

// extractHeadings counts headings by level
func (s *service) extractHeadings(doc *html.Node, analysis *WebpageAnalysis) {
	var countHeadings func(*html.Node)
	countHeadings = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				analysis.Headings[n.Data]++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			countHeadings(c)
		}
	}
	countHeadings(doc)
}

// extractLinks analyzes internal and external links
func (s *service) extractLinks(doc *html.Node, analysis *WebpageAnalysis, baseURL string) {
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
						analysis.ExternalLinks++
					} else if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") {
						analysis.InternalLinks++
					} else if strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
						// Count as external for now
						analysis.ExternalLinks++
					} else {
						// Relative links without leading slash
						analysis.InternalLinks++
					}
					break
				}
			}
			
			// Check for links without href (potentially inaccessible)
			if !hasHref {
				analysis.InaccessibleLinks++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			analyzeLinks(c)
		}
	}
	analyzeLinks(doc)
}

// extractLoginForm checks if the page contains a login form
func (s *service) extractLoginForm(doc *html.Node, analysis *WebpageAnalysis) {
	var findLoginForm func(*html.Node)
	findLoginForm = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			// Check for common login form indicators
			formText := strings.ToLower(s.getNodeText(n))
			if strings.Contains(formText, "login") || strings.Contains(formText, "sign in") || 
			   strings.Contains(formText, "username") || strings.Contains(formText, "password") ||
			   strings.Contains(formText, "email") || strings.Contains(formText, "log in") {
				analysis.HasLoginForm = true
				return
			}
			
			// Also check for input fields with login-related attributes
			var checkInputs func(*html.Node)
			checkInputs = func(node *html.Node) {
				if node.Type == html.ElementNode && node.Data == "input" {
					for _, attr := range node.Attr {
						if attr.Key == "type" {
							if attr.Val == "password" {
								analysis.HasLoginForm = true
								return
							}
						}
						if attr.Key == "name" || attr.Key == "id" {
							name := strings.ToLower(attr.Val)
							if strings.Contains(name, "user") || strings.Contains(name, "pass") ||
							   strings.Contains(name, "login") || strings.Contains(name, "email") {
								analysis.HasLoginForm = true
								return
							}
						}
					}
				}
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					checkInputs(c)
				}
			}
			checkInputs(n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLoginForm(c)
		}
	}
	findLoginForm(doc)
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
	return "Service is running and ready for webpage analysis", nil
} 