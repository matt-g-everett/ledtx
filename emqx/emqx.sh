#!/bin/bash
set -euo pipefail

# EMQX MQTT Broker Setup Script
# Idempotent - safe to run multiple times

CONTAINER_NAME="emqx"
IMAGE="emqx/emqx:latest"
DATA_VOLUME="emqx-data"
LOG_VOLUME="emqx-log"
ACL_FILE="${HOME}/emqx/acl.conf"
CONFIG_FILE="${HOME}/emqx/emqx.conf"
MQTT_PORT=1883
WS_PORT=8083
WEB_PORT=18083

# User credentials
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="${EMQX_ADMIN_PASSWORD:-change-me-admin}"
HOMEAUTO_USERNAME="homeauto"
HOMEAUTO_PASSWORD="${EMQX_HOMEAUTO_PASSWORD:-change-me-homeauto}"

echo "Setting up EMQX MQTT Broker..."

# Create/overwrite ACL file
echo "Writing ACL configuration file..."
mkdir -p "$(dirname "${ACL_FILE}")"
cat > "${ACL_FILE}" << EOF
%%--------------------------------------------------------------------
%% EMQX ACL Configuration
%%--------------------------------------------------------------------

%% Admin user - full access
{allow, {user, "${ADMIN_USERNAME}"}, subscribe, ["#"]}.
{allow, {user, "${ADMIN_USERNAME}"}, publish, ["#"]}.

%% homeauto user - restricted to home/* topics only
{allow, {user, "${HOMEAUTO_USERNAME}"}, subscribe, ["home/#"]}.
{allow, {user, "${HOMEAUTO_USERNAME}"}, publish, ["home/#"]}.

%% Deny homeauto from system topics
{deny, {user, "${HOMEAUTO_USERNAME}"}, subscribe, ["\$SYS/#"]}.
{deny, {user, "${HOMEAUTO_USERNAME}"}, publish, ["\$SYS/#"]}.

%% Deny all other access for homeauto
{deny, {user, "${HOMEAUTO_USERNAME}"}, subscribe, ["#"]}.
{deny, {user, "${HOMEAUTO_USERNAME}"}, publish, ["#"]}.

%% Default deny all for unauthenticated
{deny, all, subscribe, ["#"]}.
{deny, all, publish, ["#"]}.
EOF
echo "✓ Written ${ACL_FILE}"

echo "⚠️  WARNING: Set secure passwords via environment variables!"

# Create/overwrite EMQX config file
echo "Writing EMQX configuration..."
cat > "${CONFIG_FILE}" << EOF
# EMQX Configuration Override

# Required node settings
node {
  name = "emqx@127.0.0.1"
  cookie = "emqxsecretcookie"
  data_dir = "/opt/emqx/data"
}

# Dashboard configuration
dashboard {
  listeners.http {
    bind = "0.0.0.0:18083"
  }
  default_username = "admin"
  default_password = "public"
}

# MQTT Authentication - Built-in Database
authentication = [
  {
    mechanism = password_based
    backend = built_in_database
    user_id_type = username
    password_hash_algorithm {
      name = plain
      salt_position = disable
    }
  }
]

# Authorization (ACL)
authorization {
  sources = [
    {
      type = file
      path = "/opt/emqx/etc/acl.conf"
    }
  ]
  no_match = deny
  deny_action = ignore
  cache {
    enable = true
  }
}
EOF
echo "✓ Written ${CONFIG_FILE}"

# Create named volumes if they don't exist
if ! docker volume inspect "${DATA_VOLUME}" >/dev/null 2>&1; then
    echo "Creating data volume..."
    docker volume create "${DATA_VOLUME}"
    echo "✓ Created ${DATA_VOLUME}"
else
    echo "✓ Volume ${DATA_VOLUME} already exists"
fi

if ! docker volume inspect "${LOG_VOLUME}" >/dev/null 2>&1; then
    echo "Creating log volume..."
    docker volume create "${LOG_VOLUME}"
    echo "✓ Created ${LOG_VOLUME}"
else
    echo "✓ Volume ${LOG_VOLUME} already exists"
fi

# Pull latest image
echo "Pulling latest EMQX image..."
docker pull "${IMAGE}"

# Check if container exists
if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    echo "Container '${CONTAINER_NAME}' already exists"

    # Check if it's running
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo "Container is already running"
        echo "✓ EMQX is up and running"
    else
        echo "Starting existing container..."
        docker start "${CONTAINER_NAME}"
        echo "✓ Container started"
    fi
else
    echo "Creating new container..."
    docker run -d \
        --name "${CONTAINER_NAME}" \
        --restart always \
        -p "${MQTT_PORT}:1883" \
        -p "${WS_PORT}:8083" \
        -p "${WEB_PORT}:18083" \
        -v "${DATA_VOLUME}:/opt/emqx/data" \
        -v "${LOG_VOLUME}:/opt/emqx/log" \
        -v "${ACL_FILE}:/opt/emqx/etc/acl.conf:ro" \
        -v "${CONFIG_FILE}:/opt/emqx/etc/emqx.conf:ro" \
        "${IMAGE}"
    echo "✓ Container created and started"

    # Wait for EMQX dashboard to be ready
    echo "Waiting for EMQX dashboard to start..."
    for i in {1..60}; do
        # Check if dashboard status endpoint is available
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:${WEB_PORT}/status" 2>/dev/null || echo "000")
        if [ "${HTTP_CODE}" = "200" ]; then
            echo "✓ EMQX dashboard is responding"
            break
        fi
        echo "  Waiting... (attempt ${i}/60, HTTP ${HTTP_CODE})"
        sleep 2
    done

    # Get bearer token from dashboard (default creds are admin/public)
    echo "Getting API token..."
    LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:${WEB_PORT}/api/v5/login" \
        -H "Content-Type: application/json" \
        -d '{"username": "admin", "password": "public"}')
    echo "Login response: ${LOGIN_RESPONSE}"
    TOKEN=$(echo "${LOGIN_RESPONSE}" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

    if [ -z "${TOKEN}" ]; then
        echo "⚠ Failed to get API token"
    else
        echo "✓ Got API token"

        # Create MQTT users via API
        echo "Creating MQTT users..."

        # Create admin user
        echo "Creating admin user..."
        curl -s -X POST "http://localhost:${WEB_PORT}/api/v5/authentication/password_based%3Abuilt_in_database/users" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"user_id\": \"${ADMIN_USERNAME}\", \"password\": \"${ADMIN_PASSWORD}\", \"is_superuser\": true}"
        echo ""

        # Create homeauto user
        echo "Creating homeauto user..."
        curl -s -X POST "http://localhost:${WEB_PORT}/api/v5/authentication/password_based%3Abuilt_in_database/users" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"user_id\": \"${HOMEAUTO_USERNAME}\", \"password\": \"${HOMEAUTO_PASSWORD}\", \"is_superuser\": false}"
        echo ""

        # Change dashboard admin password
        echo "Updating dashboard admin password..."
        curl -s -X PUT "http://localhost:${WEB_PORT}/api/v5/users/admin" \
            -H "Authorization: Bearer ${TOKEN}" \
            -H "Content-Type: application/json" \
            -d "{\"password\": \"${ADMIN_PASSWORD}\", \"description\": \"Admin user\"}"
        echo ""
    fi
fi

echo ""
echo "EMQX MQTT Broker is ready!"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "MQTT Broker:   mqtt://localhost:${MQTT_PORT}"
echo "WebSocket:     ws://localhost:${WS_PORT}/mqtt"
echo "Web Dashboard: http://localhost:${WEB_PORT}"
echo ""
echo "Configuration files:"
echo "  ACL:    ${ACL_FILE}"
echo "  Config: ${CONFIG_FILE}"
echo ""
echo "Volumes:"
echo "  Data: ${DATA_VOLUME}"
echo "  Logs: ${LOG_VOLUME}"
echo ""
echo "User credentials:"
echo "  Admin:    ${ADMIN_USERNAME} / ${ADMIN_PASSWORD}"
echo "  HomeAuto: ${HOMEAUTO_USERNAME} / ${HOMEAUTO_PASSWORD}"
echo ""
echo "⚠️  To change credentials, tear down and re-run:"
echo "  docker stop ${CONTAINER_NAME}; docker rm ${CONTAINER_NAME}; docker volume rm ${DATA_VOLUME} ${LOG_VOLUME}"
echo "  ./emqx.sh"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
