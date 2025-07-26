package analyzer

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestNewHTMLParser(t *testing.T) {
	parser := NewHTMLParser()
	if parser == nil {
		t.Fatal("NewHTMLParser() returned nil")
	}
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
			if result != tt.expected {
				t.Errorf("ExtractHTMLVersion() = %v, want %v", result, tt.expected)
			}
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
			if result != tt.expected {
				t.Errorf("ExtractPageTitle() = %v, want %v", result, tt.expected)
			}
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

	if len(result) != len(expected) {
		t.Errorf("ExtractHeadings() returned %d heading types, want %d", len(result), len(expected))
	}

	for heading, count := range expected {
		if result[heading] != count {
			t.Errorf("ExtractHeadings() [%s] = %d, want %d", heading, result[heading], count)
		}
	}
}

func TestExtractLinks(t *testing.T) {
	parser := NewHTMLParser()

	tests := []struct {
		name           string
		html           string
		baseURL        string
		expectedInternal int
		expectedExternal int
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
			baseURL: "https://mysite.com",
			expectedInternal: 2,      // /internal, #anchor
			expectedExternal: 3,      // https://example.com, mailto, tel
			expectedInaccessible: 1,  // empty (javascript is filtered out)
		},
		{
			name: "No links",
			html: `<html><body><div>No links here</div></body></html>`,
			baseURL: "https://mysite.com",
			expectedInternal: 0,
			expectedExternal: 0,
			expectedInaccessible: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			internal, external, inaccessible := parser.ExtractLinks(doc, tt.baseURL)

			if internal != tt.expectedInternal {
				t.Errorf("ExtractLinks() internal = %d, want %d", internal, tt.expectedInternal)
			}
			if external != tt.expectedExternal {
				t.Errorf("ExtractLinks() external = %d, want %d", external, tt.expectedExternal)
			}
			if inaccessible != tt.expectedInaccessible {
				t.Errorf("ExtractLinks() inaccessible = %d, want %d", inaccessible, tt.expectedInaccessible)
			}
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
			name: "No forms",
			html: `<html><body><div>No forms here</div></body></html>`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			result := parser.ExtractLoginForm(doc)
			if result != tt.expected {
				t.Errorf("ExtractLoginForm() = %v, want %v", result, tt.expected)
			}
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
	if title != "Test Title" {
		t.Errorf("ExtractPageTitle() with uppercase TITLE = %v, want 'Test Title'", title)
	}

	// Test headings extraction
	headings := parser.ExtractHeadings(doc)
	if headings["h1"] != 1 {
		t.Errorf("ExtractHeadings() with uppercase H1 = %v, want 1", headings["h1"])
	}

	// Test links extraction
	internal, _, _ := parser.ExtractLinks(doc, "https://example.com")
	if internal != 1 {
		t.Errorf("ExtractLinks() with uppercase A = %v, want 1", internal)
	}

	// Test login form detection
	hasLoginForm := parser.ExtractLoginForm(doc)
	if !hasLoginForm {
		t.Error("ExtractLoginForm() with uppercase FORM/INPUT should detect login form")
	}
} 