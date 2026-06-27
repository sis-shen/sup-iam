#!/bin/sh
set -e

# ============================================================
# IAM Web Console - Docker Entrypoint
# Reads YAML config and environment variables, generates
# nginx.conf from template, then starts nginx.
# ============================================================

# Default config file path
CONFIG_FILE=${IAM_CONFIG_FILE:-/etc/iam/iam-webconsole-config.yaml}

# Default values (used as fallbacks)
SERVER_HOST="0.0.0.0"
SERVER_PORT="8080"
BACKEND_HOST="iam-api-server"
BACKEND_PORT="8080"

# Read config from YAML file if it exists
if [ -f "$CONFIG_FILE" ]; then
    echo "Reading configuration from: $CONFIG_FILE"

    # Parse server.host (supports quoted and unquoted values)
    _val=$(grep -A10 '^server:' "$CONFIG_FILE" | grep 'host:' | head -1 | sed 's/.*host: *"\(.*\)"/\1/' | sed "s/.*host: *'\(.*\)'/\1/" | sed 's/.*host: *\([^ #]*\).*/\1/')
    [ -n "$_val" ] && SERVER_HOST="$_val"

    # Parse server.port
    _val=$(grep -A10 '^server:' "$CONFIG_FILE" | grep 'port:' | head -1 | sed 's/.*port: *\([0-9]*\).*/\1/')
    [ -n "$_val" ] && SERVER_PORT="$_val"

    # Parse backend.host
    _val=$(grep -A10 '^backend:' "$CONFIG_FILE" | grep 'host:' | head -1 | sed 's/.*host: *"\(.*\)"/\1/' | sed "s/.*host: *'\(.*\)'/\1/" | sed 's/.*host: *\([^ #]*\).*/\1/')
    [ -n "$_val" ] && BACKEND_HOST="$_val"

    # Parse backend.port
    _val=$(grep -A10 '^backend:' "$CONFIG_FILE" | grep 'port:' | head -1 | sed 's/.*port: *\([0-9]*\).*/\1/')
    [ -n "$_val" ] && BACKEND_PORT="$_val"
fi

# Environment variable overrides (consistent with Viper pattern)
[ -n "$IAM_SERVER_HOST" ]   && SERVER_HOST="$IAM_SERVER_HOST"
[ -n "$IAM_SERVER_PORT" ]   && SERVER_PORT="$IAM_SERVER_PORT"
[ -n "$IAM_BACKEND_HOST" ]  && BACKEND_HOST="$IAM_BACKEND_HOST"
[ -n "$IAM_BACKEND_PORT" ]  && BACKEND_PORT="$IAM_BACKEND_PORT"

# Export for envsubst
export SERVER_HOST SERVER_PORT BACKEND_HOST BACKEND_PORT

# Log configuration
echo "Server:  ${SERVER_HOST}:${SERVER_PORT}"
echo "Backend: ${BACKEND_HOST}:${BACKEND_PORT}"

# Generate nginx.conf from template
echo "Generating nginx.conf..."
envsubst '${SERVER_HOST} ${SERVER_PORT} ${BACKEND_HOST} ${BACKEND_PORT}' \
    < /etc/nginx/templates/nginx.conf.template \
    > /etc/nginx/nginx.conf

# Verify nginx configuration
echo "Verifying nginx configuration..."
nginx -t

# Start nginx in foreground
echo "Starting nginx..."
exec nginx -g 'daemon off;'
