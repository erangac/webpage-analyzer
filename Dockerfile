# syntax=docker/dockerfile:1

# --- Build backend ---
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app

# Install basic tools
RUN apk add --no-cache git curl bash

# Copy go.mod and go.sum first for better caching
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Ensure dependencies are properly resolved
RUN go mod tidy

# Run basic Go checks
RUN go vet ./...
RUN gofmt -s -l . | grep -q . && (echo "Code formatting issues found. Run 'go fmt ./...' to fix." && exit 1) || true

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /backend ./cmd/webpage-analyzer

# --- Final image ---
FROM alpine:3.19
WORKDIR /app
COPY --from=backend-builder /backend ./backend
COPY frontend/public ./frontend/public
COPY api ./api
EXPOSE 8990
CMD ["./backend", "-port=8990"] 