# Webpage Analyzer

A Go-based tool that digs into web pages and pulls out useful information like HTML structure, links, and whether there are login forms. It's built to handle real-world websites efficiently and give you insights that matter.

## What This Tool Does

Think of it as a smart web inspector that can tell you:
- What HTML version a page uses
- The page title and heading structure
- How many links point to the same site vs external sites
- Whether there's a login form on the page
- How long the analysis took

It's particularly useful for web developers, SEO specialists, or anyone who needs to understand a website's structure quickly.

## Key Features

- **Smart Link Analysis**: Automatically categorizes links as internal, external, or broken
- **Login Form Detection**: Uses multiple strategies to spot login forms accurately
- **Parallel Processing**: Analyzes different parts of the page simultaneously for speed
- **Robust Error Handling**: Gives you clear, helpful error messages when things go wrong
- **Simple Web Interface**: A clean frontend to test the tool
- **Docker Ready**: Easy to deploy and run anywhere

## How It Works

### Link Categorization Logic

The tool uses a sophisticated approach to categorize links:

**Internal Links** (same website):
- Relative URLs like `/about`, `/contact`
- Absolute URLs pointing to the same domain (e.g., `https://example.com/page` on example.com)
- Protocol-relative URLs like `//example.com/cdn` (inherits the current page's protocol)
- Special protocols like `mailto:` and `tel:` (treated as internal)

**External Links** (different websites):
- URLs pointing to different domains
- Special protocols like `ftp://`

**Inaccessible Links** (broken or problematic):
- Empty href attributes (`href=""`)
- JavaScript links (`href="javascript:void(0)"`)
- Malformed URLs that can't be parsed

The tool uses Go's `net/url` package for robust URL parsing and domain comparison, handling edge cases like protocol-relative URLs and internationalized domain names.

### Login Form Detection

Instead of just looking for the word "login", the tool uses a multi-layered approach:

1. **Password Field Required**: First checks if there's a password input field (the strongest indicator)
2. **Form Attributes**: Looks for login-related patterns in form action, id, name, or class attributes
3. **Authentication Attributes**: Checks for autocomplete attributes like "username" or "current-password"
4. **Contextual Text**: Searches for phrases like "welcome back", "sign in to", "enter your credentials"
5. **Submit Button Text**: Looks for buttons with text like "login", "continue", "sign in"
6. **Input Field Names**: Checks for input names like "username", "user_id", "email"

This approach significantly reduces false positives (like contact forms with username fields) while catching modern login patterns that don't use obvious keywords.

### Architecture Overview

The code is organized into focused packages that each handle a specific responsibility:

```
internal/
├── analyzer/     # Main orchestration - coordinates everything
├── parser/       # HTML parsing and analysis logic
├── client/       # HTTP client for fetching web pages
├── worker/       # Parallel processing with worker pools
└── http/         # API endpoints and request handling
```

**Why This Structure?**
- **Single Responsibility**: Each package has one clear job
- **Easy Testing**: You can test each component in isolation
- **Maintainable**: Changes in one area don't break others
- **Reusable**: Components can be used in other projects

**Parallel Processing**
The tool uses a worker pool to analyze different parts of a webpage simultaneously:
- HTML version detection
- Page title extraction
- Heading analysis
- Link categorization
- Login form detection

This makes it much faster for large pages - instead of checking things one by one, it does them all at once.

## Getting Started

### Quick Start with Docker

The easiest way to get running:

```bash
# Build the Docker image (this also runs all tests)
docker build -t webpage-analyzer .

# Run it on port 8990
docker run -p 8990:8990 webpage-analyzer
```

Then open your browser to `http://localhost:8990` to see the web interface.

### Manual Setup

If you prefer to run it directly:

```bash
# Get the dependencies
go mod download

# Run the application
go run cmd/webpage-analyzer/main.go
```

The server will start on port 8080 by default.

## Using the API

### Basic Usage

Analyze a webpage with a simple POST request:

```bash
curl -X POST http://localhost:8990/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

### What You Get Back

```json
{
  "url": "https://example.com",
  "html_version": "HTML5",
  "page_title": "Example Domain",
  "headings": {
    "h1": 1,
    "h2": 3,
    "h3": 5
  },
  "internal_links": 15,
  "external_links": 8,
  "inaccessible_links": 2,
  "has_login_form": false,
  "analyzed_at": "2024-01-15T10:30:00Z",
  "processing_time": "150ms"
}
```

### Understanding the Results

- **internal_links**: Links pointing to the same website
- **external_links**: Links pointing to other websites
- **inaccessible_links**: Broken or problematic links
- **has_login_form**: Whether a login form was detected
- **processing_time**: How long the analysis took

### Error Handling

If something goes wrong, you'll get a detailed error message:

```json
{
  "status_code": 404,
  "error_message": "Not Found: The requested webpage could not be found on the server.",
  "url": "https://nonexistent.example.com"
}
```

## Testing

### Run All Tests

```bash
# Run everything with coverage
./scripts/test.sh

# Or run specific packages
go test -v ./internal/analyzer/...
go test -v ./internal/parser/...
go test -v ./internal/client/...
```

### What's Tested

- **Link categorization**: Ensures internal/external/inaccessible links are correctly identified
- **Login form detection**: Tests various login form patterns and edge cases
- **HTTP client**: Handles network errors, timeouts, and different response types
- **Parallel processing**: Verifies worker pools handle tasks correctly
- **API endpoints**: Tests all HTTP endpoints with proper error handling

## Development

### Code Quality

```bash
# Run linting and code quality checks
./scripts/lint.sh
```

This checks for:
- Code formatting
- Potential bugs
- Unused imports
- Security issues

### Building

```bash
# Build the binary
go build -o webpage-analyzer cmd/webpage-analyzer/main.go

# Run tests
go test ./...
```

## API Documentation

Once the server is running, visit `http://localhost:8990/docs` for interactive API documentation. You can test endpoints directly from your browser.

## Recent Improvements

### Link Analysis Enhancements
- **Robust URL parsing**: Uses Go's `net/url` package for reliable domain comparison
- **Protocol-relative support**: Correctly handles URLs starting with `//`
- **Special protocol handling**: Properly categorizes `mailto:`, `tel:`, and `ftp://` links
- **Internationalized domains**: Handles non-ASCII domain names correctly

### Login Form Detection Improvements
- **Multi-layered approach**: Requires password field + additional indicators
- **Modern pattern support**: Detects contemporary login forms without obvious keywords
- **Reduced false positives**: Contact forms with username fields are no longer misidentified
- **Context-aware**: Looks at surrounding text and form attributes

### Architecture Refinements
- **Package reorganization**: Separated concerns into focused packages
- **Better testability**: Each component can be tested independently
- **Cleaner interfaces**: Well-defined contracts between packages
- **Improved maintainability**: Smaller, focused files that are easier to understand

## Performance Characteristics

- **Parallel processing**: 5 concurrent tasks for faster analysis
- **Efficient memory usage**: Streams large responses without loading everything into memory
- **Timeout handling**: 30-second timeout prevents hanging on slow sites
- **Connection pooling**: Reuses HTTP connections for better performance

## Troubleshooting

### Common Issues

**"Invalid URL format"**
- Make sure the URL includes the protocol (http:// or https://)
- Check for typos in the domain name

**"Request timeout"**
- The target website might be slow or down
- Try again in a few minutes

**"Failed to parse HTML"**
- The page might not be valid HTML
- Some sites return JSON or other formats instead of HTML

### Debug Mode

For development, you can see detailed logs by setting the log level:

```bash
export LOG_LEVEL=debug
go run cmd/webpage-analyzer/main.go
```

## Contributing

This project follows standard Go conventions:
- Use `go fmt` for code formatting
- Write tests for new features
- Keep functions focused and small
- Add comments for complex logic

## License

MIT License - feel free to use this in your own projects. 