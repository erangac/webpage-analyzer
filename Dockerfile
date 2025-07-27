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

# Run comprehensive linting and testing (before generating Swaggo files)
COPY scripts/lint.sh ./scripts/lint.sh
COPY scripts/test.sh ./scripts/test.sh
RUN chmod +x ./scripts/lint.sh ./scripts/test.sh
RUN ./scripts/lint.sh
RUN ./scripts/test.sh

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