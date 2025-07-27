# Webpage Analyzer

A Go-based tool that digs into web pages and pulls out useful information like HTML structure, links, and whether there are login forms. 

## What This Tool Does

Think of it as a smart web inspector that can tell you:
- What HTML version a page uses
- The page title and heading structure
- How many links point to the same site vs external sites
- Whether there's a login form on the page
- How long the analysis took

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

**External Links** (different websites):
- URLs pointing to different domains
- Special protocols like `mailto:`, `tel:`, and `ftp://`

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

### Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Client   │    │  Web Interface  │    │   API Docs      │
│   (Browser)     │    │   (Port 8990)   │    │   (/docs)       │
└─────────┬───────┘    └─────────┬───────┘    └─────────────────┘
          │                      │
          │ HTTP Requests        │
          ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP Handlers                                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │   /health   │ │   /status   │ │  /analyze   │ │    /docs    │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘ │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Analyzer Service (Aggregator)                │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              Worker Pool (5 concurrent tasks)              │ │
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │ │
│  │  │   Parser    │ │   Client    │ │   Worker    │           │ │
│  │  │  (HTML)     │ │   (HTTP)    │ │   (Pool)    │           │ │
│  │  └─────────────┘ └─────────────┘ └─────────────┘           │ │
│  └─────────────────────────────────────────────────────────────┘ │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │              Results Aggregator                             │ │
│  │  • Collects results from all parallel tasks                │ │
│  │  • Combines into unified response                          │ │
│  │  • Adds processing time and metadata                       │ │
│  └─────────────────────────────────────────────────────────────┘ │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Analysis Results                             │
│  • HTML Version    • Page Title    • Headings                  │
│  • Internal Links  • External Links • Inaccessible Links       │
│  • Login Forms     • Processing Time                            │
└─────────────────────────────────────────────────────────────────┘
```

**How it works:**
1. **Client** sends HTTP request to analyze a webpage
2. **HTTP Handlers** receive the request and validate it
3. **Analyzer Service** orchestrates the analysis using a worker pool
4. **HTTP Client** fetches the webpage content
5. **HTML Parser** extracts structure, links, and login forms
6. **Worker Pool** runs tasks in parallel for better performance
7. **Results Aggregator** collects all parallel task results
8. **Aggregator** combines results and adds metadata (processing time, etc.)
9. **Unified response** is returned as JSON to the client

## Getting Started

### 🐳 Using Docker (Recommended)

**Docker is the easiest and most reliable way to run this tool.** It handles all dependencies and ensures consistent behavior across different environments.

```bash
# Build the Docker image (runs linting and tests during build)
docker build -t webpage-analyzer .

# Run it on port 8990
docker run -p 8990:8990 webpage-analyzer
```

Then open your browser to `http://localhost:8990` to see the web interface.

### Available URLs

Once the Docker container is running, you can access:

- **Web Interface**: `http://localhost:8990` - Interactive web UI for testing the analyzer
- **API Documentation (Swagger)**: `http://localhost:8990/docs` - Interactive API documentation with Swagger UI
- **API Endpoints**: 
  - Health check: `http://localhost:8990/api/health`
  - Analyze webpage: `http://localhost:8990/api/analyze`
  - Status: `http://localhost:8990/api/status`

### Manual Setup

If you prefer to run it directly:

```bash
# Get the dependencies
go mod download

# Run the application
go run cmd/webpage-analyzer/main.go
```

The server will start on port 8080 by default.

> **💡 Pro tip**: If you're just trying out the tool, stick with Docker. It's faster to get started, you won't need to install Go or manage dependencies, and the build process automatically runs all linting and tests to ensure code quality.

## Using the API

### Quick Test

If you're running with Docker, you can test it right away:

```bash
# Test the health endpoint
curl http://localhost:8990/api/health

# Analyze a webpage
curl -X POST http://localhost:8990/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

### Basic Usage

For more detailed analysis, use a POST request:

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

> **Note**: When building with Docker, all tests are automatically run during the build process to ensure the image only contains code that passes all checks.

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

> **Note**: Docker builds automatically run these linting checks to ensure code quality before creating the image.

### Building

```bash
# Build the binary
go build -o webpage-analyzer cmd/webpage-analyzer/main.go

# Run tests
go test ./...
```

## API Documentation

Once the server is running, visit `http://localhost:8990/docs` for interactive API documentation. You can test endpoints directly from your browser.

## Future Improvements

- **Enhanced Testing**: Add edge case testing, integration tests, and performance benchmarks
- **Dependency Injection**: Implement Wire framework for better service management
- **Code Logic**: Improve login detection, link categorization, and error handling
- **SonarQube Integration**: Implement code quality analysis with SonarQube
- **Gin Framework Integration**: Migrate to Gin framework for enhanced HTTP handling
