package analyzer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestNewHTMLParser(t *testing.T) {
	parser := NewHTMLParser()
	require.NotNil(t, parser, "NewHTMLParser() should not return nil")
}

func TestExtractHTMLVersion(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "HTML5 DOCTYPE",
			html:     "<!DOCTYPE html><html><head></head><body></body></html>",
			expected: "HTML5 (implied)",
		},
		{
			name:     "HTML4 DOCTYPE",
			html:     "<!DOCTYPE HTML PUBLIC \"-//W3C//DTD HTML 4.01//EN\"><html><head></head><body></body></html>",
			expected: "HTML4",
		},
		{
			name:     "XHTML DOCTYPE",
			html:     "<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Strict//EN\"><html><head></head><body></body></html>",
			expected: "XHTML",
		},
		{
			name:     "No DOCTYPE",
			html:     "<html><head></head><body></body></html>",
			expected: "HTML5 (implied)",
		},
		{
			name:     "Invalid document",
			html:     "invalid html",
			expected: "HTML5 (implied)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			result := parser.ExtractHTMLVersion(doc)
			assert.Equal(t, tt.expected, result, "HTML version should match expected")
		})
	}
}

func TestExtractPageTitle(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Simple title",
			html:     "<html><head><title>Test Title</title></head><body></body></html>",
			expected: "Test Title",
		},
		{
			name:     "Title with whitespace",
			html:     "<html><head><title>  Test Title  </title></head><body></body></html>",
			expected: "Test Title",
		},
		{
			name:     "No title",
			html:     "<html><head></head><body></body></html>",
			expected: "",
		},
		{
			name:     "Empty title",
			html:     "<html><head><title></title></head><body></body></html>",
			expected: "",
		},
		{
			name:     "Invalid document",
			html:     "invalid html",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			result := parser.ExtractPageTitle(doc)
			assert.Equal(t, tt.expected, result, "Page title should match expected")
		})
	}
}

func TestExtractHeadings(t *testing.T) {
	parser := NewHTMLParser()

	htmlContent := `
		<html>
			<body>
				<h1>Heading 1</h1>
				<h2>Heading 2</h2>
				<h2>Another H2</h2>
				<h3>Heading 3</h3>
				<h4>Heading 4</h4>
				<h5>Heading 5</h5>
				<h6>Heading 6</h6>
				<div>Not a heading</div>
			</body>
		</html>
	`

	doc, _ := html.Parse(strings.NewReader(htmlContent))
	result := parser.ExtractHeadings(doc)

	expected := map[string]int{
		"h1": 1,
		"h2": 2,
		"h3": 1,
		"h4": 1,
		"h5": 1,
		"h6": 1,
	}

	assert.Len(t, result, len(expected), "Number of heading types should match")
	for heading, count := range expected {
		assert.Equal(t, count, result[heading], "Heading count for %s should match", heading)
	}
}

func TestExtractLinks(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name                 string
		html                 string
		baseURL              string
		expectedInternal     int
		expectedExternal     int
		expectedInaccessible int
	}{
		{
			name: "Mixed links",
			html: `
				<html>
					<body>
						<a href="/internal">Internal</a>
						<a href="https://example.com">External</a>
						<a href="#anchor">Anchor</a>
						<a href="mailto:test@example.com">Email</a>
						<a href="tel:+1234567890">Phone</a>
						<a href="">Empty</a>
						<a href="javascript:void(0)">JavaScript</a>
					</body>
				</html>
			`,
			baseURL:              "https://mysite.com",
			expectedInternal:     2, // /internal, #anchor
			expectedExternal:     3, // https://example.com, mailto, tel
			expectedInaccessible: 1, // empty (javascript is filtered out)
		},
		{
			name:                 "No links",
			html:                 `<html><body><div>No links here</div></body></html>`,
			baseURL:              "https://mysite.com",
			expectedInternal:     0,
			expectedExternal:     0,
			expectedInaccessible: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			internal, external, inaccessible := parser.ExtractLinks(doc, tt.baseURL)

			assert.Equal(t, tt.expectedInternal, internal, "Internal links count should match")
			assert.Equal(t, tt.expectedExternal, external, "External links count should match")
			assert.Equal(t, tt.expectedInaccessible, inaccessible, "Inaccessible links count should match")
		})
	}
}

func TestExtractLoginForm(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name: "Login form with password input",
			html: `
				<html>
					<body>
						<form>
							<input type="text" name="username">
							<input type="password" name="password">
							<button type="submit">Login</button>
						</form>
					</body>
				</html>
			`,
			expected: true,
		},
		{
			name: "Login form with login keyword",
			html: `
				<html>
					<body>
						<form>
							<h2>Login Form</h2>
							<input type="text" name="user">
							<input type="password" name="pass">
							<button type="submit">Sign In</button>
						</form>
					</body>
				</html>
			`,
			expected: true,
		},
		{
			name: "Regular form without login indicators",
			html: `
				<html>
					<body>
						<form>
							<input type="text" name="name">
							<input type="text" name="comment">
							<button type="submit">Submit</button>
						</form>
					</body>
				</html>
			`,
			expected: false,
		},
		{
			name:     "No forms",
			html:     `<html><body><div>No forms here</div></body></html>`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			result := parser.ExtractLoginForm(doc)
			assert.Equal(t, tt.expected, result, "Login form detection should match expected")
		})
	}
}

func TestCaseInsensitiveElementDetection(t *testing.T) {
	parser := NewHTMLParser()

	// Test with uppercase element names
	htmlContent := `
		<html>
			<head><TITLE>Test Title</TITLE></head>
			<body>
				<H1>Heading</H1>
				<A href="/test">Link</A>
				<FORM>
					<INPUT type="password" name="pass">
				</FORM>
			</body>
		</html>
	`

	doc, _ := html.Parse(strings.NewReader(htmlContent))

	// Test title extraction
	title := parser.ExtractPageTitle(doc)
	assert.Equal(t, "Test Title", title, "Title extraction should work with uppercase TITLE")

	// Test headings extraction
	headings := parser.ExtractHeadings(doc)
	assert.Equal(t, 1, headings["h1"], "Headings extraction should work with uppercase H1")

	// Test links extraction
	internal, _, _ := parser.ExtractLinks(doc, "https://example.com")
	assert.Equal(t, 1, internal, "Links extraction should work with uppercase A")

	// Test login form detection
	hasLoginForm := parser.ExtractLoginForm(doc)
	assert.True(t, hasLoginForm, "Login form detection should work with uppercase FORM/INPUT")
}
