# Webpage Analyzer

A modern Go project for analyzing web pages with comprehensive metadata extraction, featuring a simple frontend and robust backend with parallel processing capabilities.

## 🚀 Features

- **Comprehensive Webpage Analysis**: Extract HTML version, page title, headings, links, and login forms
- **Parallel Processing**: Multi-threaded analysis using a worker pool for optimal performance
- **Robust Error Handling**: Detailed error messages for various failure scenarios
- **Simple Frontend**: Static HTML/CSS/JS interface served by the Go backend
- **Docker Support**: Easy deployment with containerization
- **OpenAPI Documentation**: Complete API specification with interactive documentation

## 📁 Project Structure

```
webpage-analyzer/
├── cmd/
│   └── webpage-analyzer/      # Main application entrypoint (main.go)
├── internal/                  # Private application code
│   ├── analyzer/              # Core webpage analysis logic
│   │   ├── html_parser.go     # HTML parsing and extraction
│   │   ├── http_client.go     # HTTP client with error handling
│   │   ├── service.go         # Main analysis service
│   │   ├── types.go           # Data structures and interfaces
│   │   └── worker_pool.go     # Parallel processing implementation
│   └── http/                  # HTTP handlers and server setup
│       └── handlers.go        # API endpoint handlers
├── frontend/
│   └── public/                # Static HTML, CSS, JS files (no build step)
├── api/                       # API specifications and documentation
│   └── openapi.yaml          # OpenAPI 3.0 specification
├── scripts/                   # Helper scripts
│   └── lint.sh               # Code quality checks
├── Dockerfile                 # Multi-stage build for backend and static frontend
├── go.mod, go.sum             # Go module files
└── README.md                  # Project documentation
```

## 🏃‍♂️ Quick Start

### Using Docker (Recommended)

Build and run with Docker:

```bash
# Build the Docker image
docker build -t webpage-analyzer .

# Run the application
docker run -p 8990:8990 webpage-analyzer
```

The application (API and static frontend) will be available at `http://localhost:8990`.

### Manual Setup

```bash
# Install dependencies
go mod download

# Run the application
go run cmd/webpage-analyzer/main.go
```

## 🌐 API Overview

The Webpage Analyzer API provides endpoints for analyzing webpages and extracting comprehensive metadata including:

- **HTML Version**: What HTML version the document uses
- **Page Title**: The title of the webpage
- **Headings**: Count of headings by level (h1, h2, h3, etc.)
- **Links**: Internal, external, and inaccessible link counts
- **Login Forms**: Detection of login forms on the page
- **Performance Metrics**: Processing time for analysis

### 🚀 Parallel Processing

The analyzer uses **multi-threaded processing** to handle large webpages efficiently:

- **5 concurrent tasks** run in parallel
- **Thread-safe operations** with proper synchronization
- **Significant performance improvement** for large pages
- **Processing time tracking** included in results

### Base URLs
- **Development**: `http://localhost:8990`
- **Production**: `https://api.webpage-analyzer.com`

## 📡 API Endpoints

### System Endpoints
- `GET /api/health` - Health check
- `GET /api/status` - Service status
- `GET /api/openapi` - OpenAPI specification (YAML)

### Analysis Endpoints
- `POST /api/analyze` - Analyze a webpage

### Documentation Endpoints
- `GET /docs` - Interactive API documentation (Swagger UI)

## 🔧 Using the API

### Health Check
```bash
curl http://localhost:8990/api/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "webpage-analyzer"
}
```

### Analyze Webpage
```bash
curl -X POST http://localhost:8990/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

**Response Example:**
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
  "inaccessible_links": 0,
  "has_login_form": false,
  "analyzed_at": "2024-01-15T10:30:00Z",
  "processing_time": "150ms"
}
```

### Error Response Example
If the URL is not reachable:
```json
{
  "status_code": 404,
  "error_message": "Not Found: The requested webpage could not be found on the server.",
  "url": "https://nonexistent.example.com"
}
```

### Get Service Status
```bash
curl http://localhost:8990/api/status
```

**Response:**
```json
{
  "status": "Service is running and ready for parallel webpage analysis with worker pool"
}
```

## 🔍 Analysis Features

### What the Analyzer Extracts:

1. **HTML Version**: Detects DOCTYPE or assumes HTML5
2. **Page Title**: Extracts content from `<title>` tag
3. **Headings**: Counts h1, h2, h3, h4, h5, h6 elements
4. **Links**: 
   - Internal links (relative URLs starting with `/` or `#`)
   - External links (absolute URLs starting with `http`)
   - Inaccessible links (links without href attributes)
5. **Login Forms**: Detects forms containing login-related keywords
6. **Processing Time**: Time taken to complete the analysis

### Parallel Processing Architecture:

The analyzer runs **5 concurrent tasks**:

1. **HTML Version Detection** - Independent task
2. **Page Title Extraction** - Independent task  
3. **Heading Analysis** - Thread-safe map operations
4. **Link Analysis** - Thread-safe counter operations
5. **Login Form Detection** - Independent boolean result

### Performance Benefits:

- **Faster processing** for large webpages
- **Better resource utilization** with concurrent tasks
- **Scalable architecture** that can handle complex pages
- **Processing time tracking** for performance monitoring

### Error Handling

The API provides detailed error messages when:
- URL is not reachable (with HTTP status code)
- Network errors occur
- HTML parsing fails
- Invalid URLs are provided

## 📚 Interactive API Documentation

### 🎯 Quick Access
Once your application is running, visit: **http://localhost:8990/docs**

This will open the interactive Swagger UI documentation where you can:
- Browse all available endpoints
- See request/response schemas
- Test API calls directly from the browser
- View examples and descriptions

### Alternative Documentation Tools

You can also view the API documentation using external tools:

1. **Using Swagger UI**: Upload `api/openapi.yaml` to [Swagger Editor](https://editor.swagger.io/)
2. **Using Redoc**: Upload `api/openapi.yaml` to [Redoc](https://redocly.github.io/redoc/)
3. **Local Development**: Use tools like `swagger-ui` or `redoc` to serve the documentation locally

## 🎨 Frontend

The frontend is a simple static site (HTML/CSS/JS) located in `frontend/public/` and is served by the Go backend. No Node.js or npm is required.

## 🛠️ Development

### Code Quality
Run the linting script to check code quality:
```bash
./scripts/lint.sh
```

### Building
```bash
# Build the application
go build -o webpage-analyzer cmd/webpage-analyzer/main.go

# Run tests
go test ./...
```

## 📋 Code Generation

The OpenAPI specification can be used to generate:
- Client SDKs (JavaScript, Python, Go, etc.)
- Server stubs
- Type definitions
- Documentation

## 🔄 Versioning

This API follows semantic versioning. The current version is `1.0.0`.

## 🚧 Future Enhancements

- Enhanced link accessibility checking
- More sophisticated login form detection
- Meta tag analysis
- Image analysis
- Performance metrics dashboard
- Authentication and authorization
- Rate limiting
- Webhook support
- Batch processing
- Real-time analysis streaming

## 📄 License

This project is open source and available under the MIT License. 