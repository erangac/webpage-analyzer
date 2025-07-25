package analyzer

import (
	"strings"

	"golang.org/x/net/html"
)

// htmlParser implements the HTMLParser interface
type htmlParser struct{}

// NewHTMLParser creates a new HTML parser instance
func NewHTMLParser() HTMLParser {
	return &htmlParser{}
}

// ExtractHTMLVersion determines the HTML version
func (p *htmlParser) ExtractHTMLVersion(doc interface{}) string {
	htmlDoc, ok := doc.(*html.Node)
	if !ok {
		return "HTML5 (implied)"
	}

	var result string
	var findDoctype func(*html.Node)
	findDoctype = func(n *html.Node) {
		if n.Type == html.DoctypeNode {
			if len(n.Attr) > 0 {
				doctype := n.Attr[0].Val
				// Handle different DOCTYPE formats
				if strings.Contains(strings.ToLower(doctype), "html5") || 
				   strings.Contains(strings.ToLower(doctype), "html 5") {
					result = "HTML5"
					return
				} else if strings.Contains(strings.ToLower(doctype), "html4") || 
				          strings.Contains(strings.ToLower(doctype), "html 4") {
					result = "HTML4"
					return
				} else if strings.Contains(strings.ToLower(doctype), "xhtml") {
					result = "XHTML"
					return
				} else {
					result = doctype
					return
				}
			} else {
				result = "HTML5 (implied)"
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findDoctype(c)
		}
	}
	findDoctype(htmlDoc)
	
	if result == "" {
		result = "HTML5 (implied)"
	}
	return result
}

// ExtractPageTitle extracts the page title
func (p *htmlParser) ExtractPageTitle(doc interface{}) string {
	htmlDoc, ok := doc.(*html.Node)
	if !ok {
		return ""
	}

	var result string
	var findTitle func(*html.Node)
	findTitle = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				result = strings.TrimSpace(n.FirstChild.Data)
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTitle(c)
		}
	}
	findTitle(htmlDoc)
	return result
}

// ExtractHeadings counts headings by level
func (p *htmlParser) ExtractHeadings(doc interface{}) map[string]int {
	htmlDoc, ok := doc.(*html.Node)
	if !ok {
		return make(map[string]int)
	}

	headings := make(map[string]int)
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
	countHeadings(htmlDoc)
	return headings
}

// ExtractLinks analyzes internal and external links
func (p *htmlParser) ExtractLinks(doc interface{}, baseURL string) (internal, external, inaccessible int) {
	htmlDoc, ok := doc.(*html.Node)
	if !ok {
		return 0, 0, 0
	}

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
	analyzeLinks(htmlDoc)
	return internal, external, inaccessible
}

// ExtractLoginForm checks if the page contains a login form
func (p *htmlParser) ExtractLoginForm(doc interface{}) bool {
	htmlDoc, ok := doc.(*html.Node)
	if !ok {
		return false
	}

	var findLoginForm func(*html.Node) bool
	findLoginForm = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "form" {
			// Check for common login form indicators
			formText := strings.ToLower(p.getNodeText(n))
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
	return findLoginForm(htmlDoc)
}

// getNodeText extracts text content from a node
func (p *htmlParser) getNodeText(n *html.Node) string {
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