#!/bin/bash
set -e

echo "Starting Enhanced Oryx..."

# Wait for Redis to be ready
echo "Waiting for Redis..."
until redis-cli ping; do
    echo "Redis is unavailable - sleeping"
    sleep 1
done
echo "Redis is up - continuing"

# Start all services
echo "Starting services with supervisor..."
exec /usr/bin/supervisord -c /etc/supervisor/conf.d/supervisord.conf


