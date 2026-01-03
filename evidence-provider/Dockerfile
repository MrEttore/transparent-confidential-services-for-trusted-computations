# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o evidenceprovider ./cmd/evidenceprovider

# Final image
FROM alpine:latest

# Install nginx
RUN apk add --no-cache nginx

WORKDIR /app

# Copy the Go binary
COPY --from=builder /app/evidenceprovider .

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

# Copy entrypoint
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Create SSL directory
RUN mkdir -p /app/ssl

# Create nginx log directory
RUN mkdir -p /var/log/nginx

EXPOSE 80 443 8080

CMD ["./entrypoint.sh"]