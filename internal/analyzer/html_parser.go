package analyzer

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// Constants for common keywords and patterns.
const (
	// HTML version keywords.
	html5Keyword = "html5"
	html4Keyword = "html4"
	xhtmlKeyword = "xhtml"

	// Login form keywords.
	loginKeyword    = "login"
	signInKeyword   = "sign in"
	usernameKeyword = "username"
	passwordKeyword = "password"
	emailKeyword    = "email"
	logInKeyword    = "log in"

	// Default values.
	defaultHTMLVersion = "HTML5 (implied)"
)

// htmlParser implements the HTMLParser interface.
type htmlParser struct{}

// NewHTMLParser creates a new HTML parser instance.
func NewHTMLParser() HTMLParser {
	return &htmlParser{}
}

// toHTMLNode safely converts interface{} to *html.Node.
func (p *htmlParser) toHTMLNode(doc interface{}) (*html.Node, bool) {
	htmlDoc, ok := doc.(*html.Node)
	return htmlDoc, ok
}

// ExtractHTMLVersion determines the HTML version.
func (p *htmlParser) ExtractHTMLVersion(doc interface{}) string {
	htmlDoc, ok := p.toHTMLNode(doc)
	if !ok {
		return defaultHTMLVersion
	}

	result := p.findDoctype(htmlDoc)
	if result == "" {
		return defaultHTMLVersion
	}
	return result
}

// findDoctype searches for DOCTYPE declaration.
func (p *htmlParser) findDoctype(n *html.Node) string {
	if n.Type == html.DoctypeNode {
		return p.parseDoctype(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := p.findDoctype(c); result != "" {
			return result
		}
	}
	return ""
}

// parseDoctype parses the DOCTYPE declaration.
func (p *htmlParser) parseDoctype(n *html.Node) string {
	if len(n.Attr) == 0 {
		return defaultHTMLVersion
	}

	doctype := strings.ToLower(n.Attr[0].Val)

	switch {
	case strings.Contains(doctype, html5Keyword) || strings.Contains(doctype, "html 5"):
		return "HTML5"
	case strings.Contains(doctype, html4Keyword) || strings.Contains(doctype, "html 4"):
		return "HTML4"
	case strings.Contains(doctype, xhtmlKeyword):
		return "XHTML"
	default:
		return n.Attr[0].Val
	}
}

// ExtractPageTitle extracts the page title.
func (p *htmlParser) ExtractPageTitle(doc interface{}) string {
	htmlDoc, ok := p.toHTMLNode(doc)
	if !ok {
		return ""
	}

	return p.findTitle(htmlDoc)
}

// findTitle searches for the title element.
func (p *htmlParser) findTitle(n *html.Node) string {
	if p.isTitleElement(n) {
		return p.extractTitleText(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := p.findTitle(c); result != "" {
			return result
		}
	}
	return ""
}

// isTitleElement checks if the node is a title element.
func (p *htmlParser) isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && strings.EqualFold(n.Data, "title")
}

// extractTitleText extracts text from title element.
func (p *htmlParser) extractTitleText(n *html.Node) string {
	if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
		return strings.TrimSpace(n.FirstChild.Data)
	}
	return ""
}

// ExtractHeadings counts headings by level.
func (p *htmlParser) ExtractHeadings(doc interface{}) map[string]int {
	htmlDoc, ok := p.toHTMLNode(doc)
	if !ok {
		return make(map[string]int)
	}

	headings := make(map[string]int)
	p.countHeadings(htmlDoc, headings)
	return headings
}

// countHeadings recursively counts heading elements.
func (p *htmlParser) countHeadings(n *html.Node, headings map[string]int) {
	if p.isHeadingElement(n) {
		headings[n.Data]++
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.countHeadings(c, headings)
	}
}

// isHeadingElement checks if the node is a heading element.
func (p *htmlParser) isHeadingElement(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	switch strings.ToLower(n.Data) {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	default:
		return false
	}
}

// ExtractLinks analyzes internal and external links.
func (p *htmlParser) ExtractLinks(doc interface{}, baseURL string) (internal, external, inaccessible int) {
	htmlDoc, ok := p.toHTMLNode(doc)
	if !ok {
		return 0, 0, 0
	}

	p.analyzeLinks(htmlDoc, baseURL, &internal, &external, &inaccessible)
	return internal, external, inaccessible
}

// analyzeLinks recursively analyzes link elements.
func (p *htmlParser) analyzeLinks(n *html.Node, baseURL string, internal, external, inaccessible *int) {
	if p.isLinkElement(n) {
		p.processLink(n, baseURL, internal, external, inaccessible)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.analyzeLinks(c, baseURL, internal, external, inaccessible)
	}
}

// isLinkElement checks if the node is a link element.
func (p *htmlParser) isLinkElement(n *html.Node) bool {
	return n.Type == html.ElementNode && strings.EqualFold(n.Data, "a")
}

// processLink processes a single link element.
func (p *htmlParser) processLink(n *html.Node, baseURL string, internal, external, inaccessible *int) {
	href := p.getHrefAttribute(n)

	if href == "" {
		*inaccessible++
		return
	}

	if !p.isValidLink(href) {
		*inaccessible++
		return
	}

	p.categorizeLink(href, baseURL, internal, external)
}

// getHrefAttribute extracts the href attribute from a link.
func (p *htmlParser) getHrefAttribute(n *html.Node) string {
	for _, attr := range n.Attr {
		if strings.EqualFold(attr.Key, "href") {
			return attr.Val
		}
	}
	return ""
}

// isValidLink checks if a link is valid (not empty or javascript).
func (p *htmlParser) isValidLink(href string) bool {
	return href != "" && !strings.HasPrefix(strings.ToLower(href), "javascript:")
}

// categorizeLink categorizes a link as internal or external based on baseURL.
func (p *htmlParser) categorizeLink(href string, baseURL string, internal, external *int) {
	// Handle relative URLs (always internal)
	if !p.isAbsoluteURL(href) {
		*internal++
		return
	}

	// Handle special protocols
	if p.isSpecialProtocol(href) {
		*external++
		return
	}

	// Compare domains for absolute URLs
	if p.isSameDomain(href, baseURL) {
		*internal++
	} else {
		*external++
	}
}

// isAbsoluteURL checks if a URL is absolute (has scheme).
func (p *htmlParser) isAbsoluteURL(href string) bool {
	return strings.HasPrefix(strings.ToLower(href), "http://") ||
		strings.HasPrefix(strings.ToLower(href), "https://") ||
		strings.HasPrefix(strings.ToLower(href), "ftp://") ||
		strings.HasPrefix(href, "//") // Protocol-relative
}

// isSpecialProtocol checks if a URL uses a special protocol.
func (p *htmlParser) isSpecialProtocol(href string) bool {
	hrefLower := strings.ToLower(href)
	return strings.HasPrefix(hrefLower, "mailto:") ||
		strings.HasPrefix(hrefLower, "tel:") ||
		strings.HasPrefix(hrefLower, "ftp://")
}

// isSameDomain checks if two URLs belong to the same domain.
func (p *htmlParser) isSameDomain(href, baseURL string) bool {
	// Handle protocol-relative URLs by adding the base scheme
	if strings.HasPrefix(href, "//") {
		baseURLParsed, err := url.Parse(baseURL)
		if err != nil {
			return false
		}
		href = baseURLParsed.Scheme + ":" + href
	}

	hrefURL, err := url.Parse(href)
	if err != nil {
		return false
	}

	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	// Compare hostnames (case-insensitive)
	return strings.EqualFold(hrefURL.Hostname(), baseURLParsed.Hostname())
}

// ExtractLoginForm checks if the page contains a login form.
func (p *htmlParser) ExtractLoginForm(doc interface{}) bool {
	htmlDoc, ok := p.toHTMLNode(doc)
	if !ok {
		return false
	}

	return p.findLoginForm(htmlDoc)
}

// findLoginForm searches for login form indicators.
func (p *htmlParser) findLoginForm(n *html.Node) bool {
	if p.isFormElement(n) {
		if p.isLoginForm(n) {
			return true
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if p.findLoginForm(c) {
			return true
		}
	}
	return false
}

// isFormElement checks if the node is a form element.
func (p *htmlParser) isFormElement(n *html.Node) bool {
	return n.Type == html.ElementNode && strings.EqualFold(n.Data, "form")
}

// isLoginForm checks if a form is a login form.
func (p *htmlParser) isLoginForm(n *html.Node) bool {
	formText := strings.ToLower(p.getNodeText(n))
	return p.containsLoginKeywords(formText) || p.hasLoginInputs(n)
}

// containsLoginKeywords checks if text contains login-related keywords.
func (p *htmlParser) containsLoginKeywords(text string) bool {
	keywords := []string{loginKeyword, signInKeyword, usernameKeyword, passwordKeyword, emailKeyword, logInKeyword}

	for _, keyword := range keywords {
		if strings.Contains(strings.ToLower(text), keyword) {
			return true
		}
	}
	return false
}

// hasLoginInputs checks if the form has login-related input fields.
func (p *htmlParser) hasLoginInputs(n *html.Node) bool {
	return p.checkInputs(n)
}

// checkInputs recursively checks for login-related input fields.
func (p *htmlParser) checkInputs(node *html.Node) bool {
	if p.isInputElement(node) {
		return p.isLoginInput(node)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if p.checkInputs(c) {
			return true
		}
	}
	return false
}

// isInputElement checks if the node is an input element.
func (p *htmlParser) isInputElement(n *html.Node) bool {
	return n.Type == html.ElementNode && strings.EqualFold(n.Data, "input")
}

// isLoginInput checks if an input field is login-related.
func (p *htmlParser) isLoginInput(n *html.Node) bool {
	for _, attr := range n.Attr {
		if p.isLoginAttribute(attr) {
			return true
		}
	}
	return false
}

// isLoginAttribute checks if an attribute indicates a login field.
func (p *htmlParser) isLoginAttribute(attr html.Attribute) bool {
	switch strings.ToLower(attr.Key) {
	case "type":
		return strings.EqualFold(attr.Val, "password")
	case "name", "id":
		return p.containsLoginKeyword(attr.Val)
	default:
		return false
	}
}

// containsLoginKeyword checks if a string contains login keywords.
func (p *htmlParser) containsLoginKeyword(s string) bool {
	name := strings.ToLower(s)
	keywords := []string{"user", "pass", "login", "email"}

	for _, keyword := range keywords {
		if strings.Contains(name, keyword) {
			return true
		}
	}
	return false
}

// getNodeText extracts text content from a node.
func (p *htmlParser) getNodeText(n *html.Node) string {
	var text strings.Builder
	p.extractText(n, &text)
	return text.String()
}

// extractText recursively extracts text from a node.
func (p *htmlParser) extractText(node *html.Node, text *strings.Builder) {
	if node.Type == html.TextNode {
		text.WriteString(node.Data)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p.extractText(c, text)
	}
}
