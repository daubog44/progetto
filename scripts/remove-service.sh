#!/bin/bash

# scripts/remove-service.sh

if [ -z "$1" ]; then
    echo "Usage: ./scripts/remove-service.sh <service-name>"
    exit 1
fi

SERVICE_NAME=$1
SERVICE_PATH="microservices/$SERVICE_NAME"

echo "🗑 Removing microservice: $SERVICE_NAME..."

# 1. Remove Directory
if [ -d "$SERVICE_PATH" ]; then
    rm -rf "$SERVICE_PATH"
    echo "✅ Deleted $SERVICE_PATH"
fi

# 2. Update go.work
sed -i "\|\./microservices/$SERVICE_NAME|d" go.work
echo "✅ Updated go.work"

# 3. Update docker-compose.yml
# Removes the service block from docker-compose.yml
# It looks for the service name and deletes until it finds a line with fewer than 2 spaces (like 'volumes:') or end of file
sed -i "/  $SERVICE_NAME:/,/^[[:alnum:]]/ { /^[[:alnum:]]/!d }" docker-compose.yml
echo "✅ Updated docker-compose.yml"

# 4. Update Tiltfile
sed -i "/# $SERVICE_NAME/,+11d" Tiltfile
echo "✅ Updated Tiltfile"

echo "🎉 Service $SERVICE_NAME removed successfully!"
