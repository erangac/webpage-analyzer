# API Documentation

This folder contains API definitions, contracts, and documentation for the Webpage Analyzer service.

## Contents

- **`openapi.yaml`** - OpenAPI 3.0 specification defining all API endpoints
- **`README.md`** - This documentation file

## API Overview

The Webpage Analyzer API provides endpoints for analyzing webpages and extracting specific metadata including:

- **HTML Version**: What HTML version the document uses
- **Page Title**: The title of the webpage
- **Headings**: Count of headings by level (h1, h2, h3, etc.)
- **Links**: Internal and external link counts
- **Login Forms**: Detection of login forms on the page

### Base URLs
- **Development**: `http://localhost:8990`
- **Production**: `https://api.webpage-analyzer.com`

### Endpoints

#### System Endpoints
- `GET /api/health` - Health check
- `GET /api/status` - Service status

#### Analysis Endpoints
- `POST /api/analyze` - Analyze a webpage

#### Documentation Endpoints
- `GET /docs` - Interactive API documentation (Swagger UI)
- `GET /api/openapi` - OpenAPI specification (YAML)

## Using the API

### Health Check
```bash
curl http://localhost:8990/api/health
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
  "html_version": "HTML5 (implied)",
  "page_title": "Example Domain",
  "headings": {
    "h1": 1,
    "h2": 3
  },
  "internal_links": 5,
  "external_links": 2,
  "inaccessible_links": 0,
  "has_login_form": false,
  "analyzed_at": "2024-01-15T10:30:00Z"
}
```

### Error Response Example
If the URL is not reachable:
```json
{
  "status_code": 404,
  "error_message": "HTTP 404: Not Found",
  "url": "https://nonexistent.example.com"
}
```

### Get Service Status
```bash
curl http://localhost:8990/api/status
```

## Analysis Features

### What the Analyzer Extracts:

1. **HTML Version**: Detects DOCTYPE or assumes HTML5
2. **Page Title**: Extracts content from `<title>` tag
3. **Headings**: Counts h1, h2, h3, h4, h5, h6 elements
4. **Links**: 
   - Internal links (relative URLs starting with `/` or `#`)
   - External links (absolute URLs starting with `http`)
   - Inaccessible links (currently set to 0, can be enhanced)
5. **Login Forms**: Detects forms containing login-related keywords

### Error Handling

The API provides detailed error messages when:
- URL is not reachable (with HTTP status code)
- Network errors occur
- HTML parsing fails
- Invalid URLs are provided

## Interactive API Documentation

### 🎯 **Quick Access**
Once your application is running, visit: **http://localhost:8990/docs**

This will open the interactive Swagger UI documentation where you can:
- Browse all available endpoints
- See request/response schemas
- Test API calls directly from the browser
- View examples and descriptions

### Alternative Documentation Tools

You can also view the API documentation using external tools:

1. **Using Swagger UI**: Upload `openapi.yaml` to [Swagger Editor](https://editor.swagger.io/)
2. **Using Redoc**: Upload `openapi.yaml` to [Redoc](https://redocly.github.io/redoc/)
3. **Local Development**: Use tools like `swagger-ui` or `redoc` to serve the documentation locally

## Code Generation

The OpenAPI specification can be used to generate:
- Client SDKs (JavaScript, Python, Go, etc.)
- Server stubs
- Type definitions
- Documentation

## Versioning

This API follows semantic versioning. The current version is `1.0.0`.

## Future Enhancements

- Enhanced link accessibility checking
- More sophisticated login form detection
- Meta tag analysis
- Image analysis
- Performance metrics
- Authentication and authorization
- Rate limiting
- Webhook support
- Batch processing 