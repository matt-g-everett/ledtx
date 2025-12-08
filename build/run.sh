#!/bin/bash
set -e

# Get the project root directory
SCRIPT_DIR="$(dirname "$0")"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Running ledtx container..."
echo "Config: $PROJECT_ROOT/homeauto.yaml -> /app/config.yaml"

docker run --rm -it \
  -v "$PROJECT_ROOT/homeauto.yaml:/app/config.yaml:ro" \
  --network host \
  ledtx:latest
