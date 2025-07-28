# syntax=docker/dockerfile:1

# --- Build backend ---
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app

# Install basic tools and Swaggo
RUN apk add --no-cache git curl bash dos2unix
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy scripts first (since they're excluded in .dockerignore)
COPY scripts/ ./scripts/

# Copy source code (excluding unnecessary files)
COPY . .

# Ensure scripts have proper line endings and permissions
RUN dos2unix scripts/lint.sh scripts/test.sh 2>/dev/null || true
RUN chmod +x scripts/lint.sh scripts/test.sh

# Verify both scripts exist and are executable
RUN ls -la scripts/ && echo "Scripts found and ready"
RUN test -f scripts/lint.sh && echo "lint.sh exists" || echo "ERROR: lint.sh not found"
RUN test -f scripts/test.sh && echo "test.sh exists" || echo "ERROR: test.sh not found"
RUN test -x scripts/lint.sh && echo "lint.sh is executable" || echo "ERROR: lint.sh not executable"
RUN test -x scripts/test.sh && echo "test.sh is executable" || echo "ERROR: test.sh not executable"

# Run comprehensive linting and testing (before generating Swaggo files)
RUN bash scripts/lint.sh
RUN bash scripts/test.sh

# Generate OpenAPI specification using Swaggo
RUN swag init -g cmd/webpage-analyzer/main.go -o api

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