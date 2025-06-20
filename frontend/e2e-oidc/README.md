# Nebraska OIDC E2E Test Suite

This directory contains end-to-end tests specifically for testing the OIDC (OpenID Connect) authentication and authorization implementation in Nebraska.

## Overview

The OIDC test suite is completely separate from the main E2E tests and includes:

- **Separate infrastructure**: Uses different ports, database, and Docker containers
- **Keycloak integration**: Tests against a real Keycloak instance with pre-configured realm
- **Comprehensive coverage**: Authentication flows, authorization levels, token handling, and error scenarios

## Test Structure

```
e2e-oidc/
├── oidc-global.setup.ts           # Test environment setup
├── oidc-global.teardown.ts        # Test environment cleanup
├── auth-flows.spec.ts             # Authentication flow tests
├── authorization-admin.spec.ts    # Admin role authorization tests
├── authorization-viewer.spec.ts   # Viewer role authorization tests
├── token-expiration.spec.ts       # Token expiration handling
├── invalid-tokens.spec.ts         # Invalid token scenarios
├── api-protection.spec.ts         # API endpoint protection
├── mask-oidc-dynamic-parts.css    # CSS for consistent screenshots
└── helpers/
    ├── oidc-helpers.ts            # Main test utilities
    ├── token-manager.ts           # JWT token manipulation
    ├── keycloak-api.ts           # Keycloak API interactions
    └── test-users.ts             # Test user definitions
```

## Test Users

The test suite uses these pre-configured users:

- **test-admin** (password: admin123)
  - Roles: `test_admin`, `test_viewer`
  - Full read/write access to all endpoints

- **test-viewer** (password: viewer123)
  - Roles: `test_viewer`
  - Read-only access, write operations return 403

- **test-expired** (password: expired123)
  - Used for testing token expiration scenarios

## Running Tests

### Prerequisites

1. Docker/Podman for running containers
2. Node.js and npm dependencies installed

### Commands

```bash
# Run all OIDC tests
npm run test:oidc

# Run with UI (interactive mode)
npm run test:oidc:ui

# Run in headed mode (visible browser)
npm run test:oidc:headed

# Run with debugging
npm run test:oidc:debug
```

### Local Development

For local development without CI:

```bash
# Start the OIDC test environment manually
cd ../backend
docker compose -f docker-compose.base.yaml -f docker-compose.oidc-test.yaml up --build

# In another terminal, run tests
npm run test:oidc
```

## Infrastructure

### Docker Compose

The OIDC tests use `docker-compose.base.yaml` and `docker-compose.oidc-test.yaml` which includes:

- **PostgreSQL** (port 8001): Test database
- **Keycloak** (port 8063): OIDC provider with test realm
- **Nebraska Backend** (port 8003): With OIDC authentication enabled

### Keycloak Configuration

- **Realm**: `test`
- **Client ID**: `nebraska-test`
- **Client Type**: Public (no client secret required)
- **Grant Types**: Authorization Code, Direct Access (for testing)
- **Redirect URIs**: `http://localhost:8003/*`, `http://127.0.0.1:8003/*`

## Test Coverage

### Authentication Tests (`auth-flows.spec.ts`)
- Token acquisition for different user types
- OIDC discovery endpoint validation
- JWT token structure validation
- Concurrent authentication requests
- Keycloak connectivity handling

### Authorization Tests (`authorization-admin.spec.ts`, `authorization-viewer.spec.ts`)
- Role-based access control (RBAC)
- Admin vs viewer permission differences
- API endpoint access patterns
- Write operation restrictions for viewers

### Token Handling (`token-expiration.spec.ts`, `invalid-tokens.spec.ts`)
- Expired token rejection
- Invalid token formats
- Malformed JWT handling
- Token signature validation
- Missing claims handling

### API Protection (`api-protection.spec.ts`)
- Endpoint protection without authentication
- Concurrent authenticated requests
- Different HTTP methods (GET, POST, PUT, DELETE)
- Query parameters and large payloads

## Key Features

### Token Expiration Testing

The test suite simulates token expiration scenarios without waiting:

- **Mock expired tokens**: Creates tokens with past expiration times
- **Short-lived tokens**: Uses tokens with very short lifespans
- **Expiration validation**: Tests backend rejection of expired tokens

### Role-Based Testing

Comprehensive testing of the two-tier permission system:

- **Admin users**: Full CRUD access to all resources
- **Viewer users**: Read-only access, write operations blocked with 403

### Error Scenario Coverage

Tests various error conditions:

- **Authentication errors**: Invalid/missing tokens (401)
- **Authorization errors**: Insufficient permissions (403)
- **Malformed requests**: Invalid JWT structure, wrong algorithms
- **Network conditions**: Concurrent requests, large payloads

## Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8001, 8003, and 8063 are available
2. **Container startup**: Keycloak takes time to start, tests wait automatically
3. **Database setup**: Tests automatically populate required data

### Debug Mode

Use debug mode to step through tests:

```bash
npm run test:oidc:debug
```

### Logs

View container logs:

```bash
# Keycloak logs
docker compose -f docker-compose.base.yaml -f docker-compose.oidc-test.yaml logs keycloak

# Nebraska backend logs
docker compose -f docker-compose.base.yaml -f docker-compose.oidc-test.yaml logs server
```

## CI/CD Integration

The test suite is designed for CI environments:

- **Automatic container management**: Docker Compose handles startup/teardown
- **Environment detection**: Uses different ports/URLs for CI vs local
- **Timeout handling**: Generous timeouts for container startup
- **Parallel execution**: Tests can run in parallel where appropriate

## Extending Tests

To add new OIDC tests:

1. Create new `.spec.ts` files in the `e2e-oidc/` directory
2. Use the existing helpers from `helpers/` directory
3. Follow the naming convention: `feature-name.spec.ts`
4. Add appropriate test data to `test-users.ts` if needed

### Helper Functions

The `OIDCHelpers` class provides utilities for:

- **Token management**: Getting valid/expired/invalid tokens
- **API testing**: Making authenticated requests
- **Role testing**: Testing different permission levels
- **Error simulation**: Creating various error scenarios

## Architecture Notes

This test suite validates the OIDC refactoring that moved Nebraska from session-based authentication to stateless JWT validation:

- **No sessions**: Backend validates JWTs using public key cryptography
- **Role-based authorization**: Uses realm roles from Keycloak tokens
- **Stateless operation**: No server-side session storage
- **Public client**: Frontend uses authorization code + PKCE flow (simulated in tests)