FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN go build -o evidenceprovider ./cmd/evidenceprovider

FROM alpine:latest

RUN apk add --no-cache nginx

WORKDIR /app

COPY --from=builder /app/evidenceprovider .

COPY nginx.conf /etc/nginx/nginx.conf

COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

RUN mkdir -p /app/ssl

RUN mkdir -p /var/log/nginx

EXPOSE 80 443 8080

CMD ["./entrypoint.sh"]