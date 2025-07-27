package parser

// HTMLParser defines the interface for HTML parsing operations.
type HTMLParser interface {
	ExtractHTMLVersion(doc interface{}) string
	ExtractPageTitle(doc interface{}) string
	ExtractHeadings(doc interface{}) map[string]int
	ExtractLinks(doc interface{}, baseURL string) (internal, external, inaccessible int)
	ExtractLoginForm(doc interface{}) bool
}
