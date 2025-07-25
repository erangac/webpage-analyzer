# API Documentation

This folder contains API definitions, contracts, and documentation for the Webpage Analyzer service.

## Contents

- **`openapi.yaml`** - OpenAPI 3.0 specification defining all API endpoints
- **`README.md`** - This documentation file

## API Overview

The Webpage Analyzer API provides endpoints for analyzing webpages and extracting metadata.

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

### Get Service Status
```bash
curl http://localhost:8990/api/status
```

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

- API versioning (v1, v2, etc.)
- Authentication and authorization
- Rate limiting
- Webhook support
- Batch processing 