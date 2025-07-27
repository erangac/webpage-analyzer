package parser

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

	// Login form keywords (more specific to reduce false positives).
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

// isLoginForm checks if a form is a login form using a more robust approach.
func (p *htmlParser) isLoginForm(n *html.Node) bool {
	// 1. Check for password input (strongest indicator)
	hasPassword := p.hasPasswordInput(n)
	if !hasPassword {
		return false // No password field = not a login form
	}

	// 2. Check for login-specific patterns in form attributes
	hasLoginPattern := p.hasLoginPattern(n)

	// 3. Check for authentication-related attributes
	hasAuthAttributes := p.hasAuthAttributes(n)

	// 4. Check for login-related text (but be more specific)
	hasLoginText := p.hasSpecificLoginText(n)

	// 5. Check for submit button with login text
	hasLoginSubmit := p.hasLoginSubmitButton(n)

	// 6. Check for login-related input names/ids
	hasLoginInputs := p.hasLoginInputs(n)

	// If password field exists, require at least one other indicator
	// This is more permissive than requiring multiple indicators but still more specific than the original
	return hasPassword && (hasLoginPattern || hasAuthAttributes || hasLoginText || hasLoginSubmit || hasLoginInputs)
}

// hasPasswordInput checks if the form contains a password input field.
func (p *htmlParser) hasPasswordInput(n *html.Node) bool {
	return p.hasInputWithType(n, "password")
}

// hasInputWithType checks if the form contains an input with the specified type.
func (p *htmlParser) hasInputWithType(node *html.Node, inputType string) bool {
	if p.isInputElement(node) {
		for _, attr := range node.Attr {
			if strings.EqualFold(attr.Key, "type") && strings.EqualFold(attr.Val, inputType) {
				return true
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if p.hasInputWithType(c, inputType) {
			return true
		}
	}
	return false
}

// hasLoginPattern checks for login-specific patterns in form attributes.
func (p *htmlParser) hasLoginPattern(n *html.Node) bool {
	// Check form action, id, name, class for login patterns
	for _, attr := range n.Attr {
		attrValue := strings.ToLower(attr.Val)
		switch strings.ToLower(attr.Key) {
		case "action", "id", "name", "class":
			if p.containsLoginPattern(attrValue) {
				return true
			}
		}
	}
	return false
}

// containsLoginPattern checks if a string contains login-specific patterns.
func (p *htmlParser) containsLoginPattern(s string) bool {
	patterns := []string{
		"login", "signin", "sign_in", "sign-in",
		"authenticate", "auth", "authentication",
		"logon", "signon", "sign_on", "sign-on",
	}

	for _, pattern := range patterns {
		if strings.Contains(s, pattern) {
			return true
		}
	}
	return false
}

// hasAuthAttributes checks for authentication-related attributes.
func (p *htmlParser) hasAuthAttributes(n *html.Node) bool {
	// Check for common authentication-related attributes
	authAttrs := []string{"autocomplete", "data-auth", "data-login"}

	for _, attr := range n.Attr {
		attrKey := strings.ToLower(attr.Key)
		attrValue := strings.ToLower(attr.Val)

		for _, authAttr := range authAttrs {
			if strings.Contains(attrKey, authAttr) {
				return true
			}
		}

		// Check for autocomplete values related to login
		if attrKey == "autocomplete" {
			loginAutocomplete := []string{"username", "current-password", "new-password"}
			for _, loginAuto := range loginAutocomplete {
				if strings.Contains(attrValue, loginAuto) {
					return true
				}
			}
		}
	}
	return false
}

// hasSpecificLoginText checks for login-related text with more specific patterns.
func (p *htmlParser) hasSpecificLoginText(n *html.Node) bool {
	formText := strings.ToLower(p.getNodeText(n))

	// More specific login-related phrases
	loginPhrases := []string{
		"sign in to", "log in to", "login to",
		"welcome back", "welcome to",
		"enter your", "provide your",
		"access your account", "access account",
		"your credentials", "your password",
		"authentication required", "login required",
	}

	for _, phrase := range loginPhrases {
		if strings.Contains(formText, phrase) {
			return true
		}
	}

	// Check for specific input labels
	inputLabels := []string{
		"username", "user id", "userid", "user-id",
		"email address", "email addr", "e-mail",
		"password", "passwd", "pass word", "pass-word",
	}

	for _, label := range inputLabels {
		if strings.Contains(formText, label) {
			return true
		}
	}

	return false
}

// hasLoginSubmitButton checks for submit buttons with login-related text.
func (p *htmlParser) hasLoginSubmitButton(n *html.Node) bool {
	return p.findSubmitButtonWithText(n, []string{
		"login", "sign in", "signin", "log in",
		"authenticate", "continue", "submit",
		"enter", "access", "proceed",
	})
}

// findSubmitButtonWithText checks for submit buttons with specific text.
func (p *htmlParser) findSubmitButtonWithText(node *html.Node, buttonTexts []string) bool {
	if p.isSubmitButton(node) {
		buttonText := strings.ToLower(p.getNodeText(node))
		for _, text := range buttonTexts {
			if strings.Contains(buttonText, text) {
				return true
			}
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if p.findSubmitButtonWithText(c, buttonTexts) {
			return true
		}
	}
	return false
}

// isSubmitButton checks if the node is a submit button.
func (p *htmlParser) isSubmitButton(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	switch strings.ToLower(n.Data) {
	case "button", "input":
		for _, attr := range n.Attr {
			if strings.EqualFold(attr.Key, "type") && strings.EqualFold(attr.Val, "submit") {
				return true
			}
		}
	}
	return false
}

// containsLoginKeywords checks if text contains login-related keywords.
// This is kept for backward compatibility but is now more specific.
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
	// More specific keywords to reduce false positives
	keywords := []string{"username", "userid", "user_id", "user-name", "password", "passwd", "pass_word", "pass-word", "login", "email"}

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
