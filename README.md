# Webpage Analyzer

A modern Go project for analyzing web pages, with a simple frontend and backend.

## Folder Structure

```
webpage-analyzer/
├── cmd/
│   └── webpage-analyzer/      # Main application entrypoint (main.go)
├── internal/                  # Private application code
│   ├── analyzer/              # Core webpage analysis logic
│   ├── http/                  # HTTP handlers and server setup
│   └── utils/                 # Utility packages
├── pkg/                       # Public Go packages (optional, for reusable code)
├── frontend/
│   └── public/                # Static HTML, CSS, JS files (no build step)
├── api/                       # OpenAPI specs, API contracts, or protobufs (optional)
├── scripts/                   # Helper scripts (build, test, etc.)
├── test/                      # Integration and end-to-end tests
├── Dockerfile                 # Multi-stage build for backend and static frontend
├── go.mod, go.sum             # Go module files
└── README.md                  # Project documentation
```

## Usage

Build and run with Docker:

```sh
docker build -t webpage-analyzer .
docker run -p 8990:8990 webpage-analyzer
```

The application (API and static frontend) will be available at `http://localhost:8990`.

## Frontend

The frontend is a simple static site (HTML/CSS/JS) located in `frontend/public/` and is served by the Go backend. No Node.js or npm is required. 