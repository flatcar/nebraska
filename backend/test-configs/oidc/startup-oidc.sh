#!/bin/sh
set -e

echo "Starting Nebraska OIDC startup script..."

# Function to wait for postgres
wait_for_postgres() {
    echo "Waiting for PostgreSQL to be ready..."
    until nc -z localhost 8001; do
        echo "PostgreSQL not ready on localhost:8001, waiting..."
        sleep 2
    done
    echo "PostgreSQL is ready!"
}

# Function to wait for keycloak
wait_for_keycloak() {
    echo "Waiting for Keycloak to be ready..."
    until curl -f -s http://localhost:8063/realms/test > /dev/null 2>&1; do
        echo "Keycloak not ready at localhost:8063, waiting..."
        sleep 5
    done
    echo "Keycloak is ready!"
}

# Wait for dependencies
wait_for_postgres
wait_for_keycloak

echo "All dependencies are ready. Starting Nebraska backend..."

# Start Nebraska with OIDC configuration
exec /nebraska/nebraska \
    --auth-mode=oidc \
    --oidc-issuer-url=http://127.0.0.1:8063/realms/test \
    --oidc-client-id=nebraska-test \
    --oidc-admin-roles=test_admin \
    --oidc-viewer-roles=test_viewer \
    --oidc-roles-path=realm_access.roles \
    --oidc-scopes=openid,profile,email \
    --http-static-dir=/nebraska/static \
    --api-endpoint-suffix=/ \
    --port=8003
