# syntax=docker/dockerfile:1

# --- Build backend ---
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /backend ./cmd/webpage-analyzer

# --- Final image ---
FROM alpine:3.19
WORKDIR /app
COPY --from=backend-builder /backend ./backend
COPY frontend/public ./frontend/public
EXPOSE 8990
CMD ["./backend", "-port=8990"] 