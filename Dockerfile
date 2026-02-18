FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o dukapos ./cmd/server/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

# Copy binary from builder
COPY --from=builder /app/dukapos .

# Create logs directory
RUN mkdir -p /app/logs

# Copy .env.example as template
COPY --from=builder /app/.env.example .env.example

# Copy static files if they exist
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Expose port
EXPOSE 8080

# Run the application
CMD ["./dukapos"]
