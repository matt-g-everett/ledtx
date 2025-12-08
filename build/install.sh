#!/bin/bash
set -e

# Get the project root directory
SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

CONTAINER_NAME="ledtx"

# Stop and remove existing container if it exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Stopping and removing existing container: ${CONTAINER_NAME}"
    docker stop "${CONTAINER_NAME}" || true
    docker rm "${CONTAINER_NAME}" || true
fi

echo "Installing ledtx container as a persistent service..."
echo "Config: $PROJECT_ROOT/homeauto.yaml -> /app/config.yaml"

docker run -d \
  --name "${CONTAINER_NAME}" \
  --restart unless-stopped \
  -v "$PROJECT_ROOT/homeauto.yaml:/app/config.yaml:ro" \
  --network host \
  ledtx:latest

echo "Container ${CONTAINER_NAME} installed and running"
echo "View logs with: docker logs -f ${CONTAINER_NAME}"
echo "Stop with: docker stop ${CONTAINER_NAME}"
