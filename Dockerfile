# syntax=docker/dockerfile:1

# --- Build backend ---
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app

# Install static analysis tools
RUN apk add --no-cache git curl bash
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2
RUN go install honnef.co/go/tools/cmd/staticcheck@v0.4.6

# Copy go.mod and go.sum first for better caching
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Ensure dependencies are properly resolved
RUN go mod tidy

# Run static analysis
RUN golangci-lint run --timeout=5m
RUN go vet ./...
RUN gofmt -s -l . | grep -q . && (echo "Code formatting issues found. Run 'go fmt ./...' to fix." && exit 1) || true
RUN staticcheck ./...

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