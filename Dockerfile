# syntax=docker/dockerfile:1

# --- Build backend ---
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app

# Install basic tools and Swaggo
RUN apk add --no-cache git curl bash
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go.mod and go.sum first for better caching
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Generate OpenAPI specification using Swaggo
RUN swag init -g cmd/webpage-analyzer/main.go -o api

# Ensure dependencies are properly resolved
RUN go mod tidy

    # Run unit tests with coverage
    RUN go test -v -coverprofile=coverage.out -covermode=atomic ./internal/analyzer/... -timeout=30s
    RUN go test -v -coverprofile=coverage.out -covermode=atomic ./internal/http/... -timeout=30s
    RUN go test -v -coverprofile=coverage.out -covermode=atomic ./cmd/webpage-analyzer/... -timeout=30s

# Display test coverage summary
RUN go tool cover -func=coverage.out

# Run basic Go checks
RUN go vet ./...
RUN gofmt -s -l . | grep -q . && (echo "Code formatting issues found. Run 'go fmt ./...' to fix." && exit 1) || true

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /backend ./cmd/webpage-analyzer

# --- Final image ---
FROM alpine:3.19
WORKDIR /app
COPY --from=backend-builder /backend ./backend
COPY --from=backend-builder /app/api ./api
COPY frontend/public ./frontend/public
EXPOSE 8990
CMD ["./backend", "-port=8990"] 