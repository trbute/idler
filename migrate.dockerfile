FROM golang:1.24.1-alpine

# Install goose once and cache it
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Set working directory
WORKDIR /migrations

# Default command - can be overridden
CMD ["goose", "up"]