# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o forum ./main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache sqlite-libs ca-certificates

# Create app directories
RUN mkdir -p /app/BackEnd/database/storage \
    && mkdir -p /app/FrontEnd/templates \
    && mkdir -p /app/FrontEnd/static \
    && chmod 755 /app/BackEnd/database/storage

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/forum .

# Copy templates and static files
COPY FrontEnd/templates ./FrontEnd/templates
COPY FrontEnd/static ./FrontEnd/static

# Create a named volume for database persistence
VOLUME /app/BackEnd/database/storage

# Expose port
EXPOSE 8080

# Run the application
CMD ["./forum"]