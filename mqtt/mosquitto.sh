#!/bin/bash
set -euo pipefail

# Mosquitto MQTT Broker Setup Script
# Ultra-simple, stateless broker with no security - for development/testing only!
# Idempotent - safe to run multiple times

CONTAINER_NAME="mosquitto"
IMAGE="eclipse-mosquitto:latest"
MQTT_PORT=1883

# User credentials
HOMEAUTO_USERNAME="homeauto"
HOMEAUTO_PASSWORD="${MQTT_HOMEAUTO_PASSWORD:-change-me-homeauto}"

echo "Setting up Mosquitto MQTT Broker..."
echo "⚠️  WARNING: This broker uses simple password authentication."
echo "    For development/testing purposes."
echo ""

# Pull latest image
echo "Pulling latest Mosquitto image..."
docker pull "${IMAGE}"

# Check if container exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Container '${CONTAINER_NAME}' already exists"

    # Check if it's running
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo "Container is already running"
        echo "✓ Mosquitto is up and running"
    else
        echo "Starting existing container..."
        docker start "${CONTAINER_NAME}"
        echo "✓ Container started"
    fi
else
    echo "Creating new container with authentication..."
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --restart unless-stopped \
        -p "${MQTT_PORT}:1883" \
        "${IMAGE}" \
        sh -c "echo 'listener 1883' > /mosquitto/config/mosquitto.conf && \
               echo 'allow_anonymous false' >> /mosquitto/config/mosquitto.conf && \
               echo 'password_file /mosquitto/config/password.txt' >> /mosquitto/config/mosquitto.conf && \
               mosquitto_passwd -c -b /mosquitto/config/password.txt ${HOMEAUTO_USERNAME} ${HOMEAUTO_PASSWORD} && \
               chown mosquitto:mosquitto /mosquitto/config/password.txt && \
               chmod 600 /mosquitto/config/password.txt && \
               mosquitto -c /mosquitto/config/mosquitto.conf"
    echo "✓ Container created and started"

    # Brief wait for startup
    echo "Waiting for Mosquitto to start..."
    sleep 2
fi

echo ""
echo "Mosquitto MQTT Broker is ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "MQTT Broker:   mqtt://localhost:${MQTT_PORT}"
echo ""
echo "Configuration: Password authentication enabled"
echo "State:         Stateless (no volumes)"
echo ""
echo "User credentials:"
echo "  Username: ${HOMEAUTO_USERNAME}"
echo "  Password: ${HOMEAUTO_PASSWORD}"
echo ""
echo "⚠️  To change credentials, set MQTT_HOMEAUTO_PASSWORD environment variable"
echo "    and restart: docker stop ${CONTAINER_NAME} && docker rm ${CONTAINER_NAME} && ./mosquitto.sh"
echo ""
echo "To stop and remove:"
echo "  docker stop ${CONTAINER_NAME} && docker rm ${CONTAINER_NAME}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
