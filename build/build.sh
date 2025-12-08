#!/bin/bash
set -e

# Change to project root directory
cd "$(dirname "$0")/.."

echo "Building ledtx container with BuildKit cache..."

DOCKER_BUILDKIT=1 docker build \
  -f build/Dockerfile \
  -t ledtx:latest \
  .

echo "Build complete! Image tagged as ledtx:latest"
