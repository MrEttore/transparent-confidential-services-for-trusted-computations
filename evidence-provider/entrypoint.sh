#!/bin/sh
set -e

echo "Starting Evidence Provider..."

# Start the Go application (it generates TLS cert at startup).
./evidenceprovider &
GO_PID=$!

# Wait for certificate to be generated.
echo "Waiting for TLS certificate..."
for i in $(seq 1 30); do
    if [ -f /app/ssl/cert.pem ] && [ -f /app/ssl/key.pem ]; then
        echo "TLS certificate ready!"
        break
    fi
    sleep 0.5
done

if [ ! -f /app/ssl/cert.pem ]; then
    echo "TLS certificate not generated after 15s"
    exit 1
fi

# Start nginx.
echo "Starting nginx..."
nginx -g 'daemon off;' &
NGINX_PID=$!

wait -n $GO_PID $NGINX_PID

kill $GO_PID $NGINX_PID 2>/dev/null || true
