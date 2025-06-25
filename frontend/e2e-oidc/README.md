# Nebraska OIDC E2E Tests

End-to-end tests for Nebraska's OIDC authentication and authorization implementation.

## Test Structure

```
e2e-oidc/
├── api-protection.spec.ts         # API endpoint protection
├── auth-flows.spec.ts            # Authentication flow tests
├── authorization-admin.spec.ts   # Admin role tests
├── authorization-viewer.spec.ts  # Viewer role tests
├── invalid-tokens.spec.ts        # Invalid token handling
├── token-expiration.spec.ts      # Token expiration tests
├── test-config.ts               # Centralized configuration
└── helpers/                     # Test utilities
```

## Running Tests

```bash
# Run all OIDC tests
npm run test:oidc

# Run with UI (interactive mode)
npm run test:oidc:ui

# Run in headed mode (visible browser)
npm run test:oidc:headed

# Debug mode
npm run test:oidc:debug
```

## Test Infrastructure

Uses Docker Compose with:

- **PostgreSQL** (port 8001): Test database
- **Keycloak** (port 8063): OIDC provider
- **Nebraska Backend** (port 8003): OIDC-enabled server

## Test Users

| User        | Password  | Roles                   | Access          |
| ----------- | --------- | ----------------------- | --------------- |
| test-admin  | admin123  | test_admin, test_viewer | Full read/write |
| test-viewer | viewer123 | test_viewer             | Read-only       |

## Test Coverage

- **Authentication**: Token acquisition, OIDC discovery, JWT validation
- **Authorization**: Role-based access control, admin vs viewer permissions
- **Token Handling**: Expired tokens, invalid formats, malformed JWTs
- **API Protection**: Unauthenticated access, HTTP methods, nested endpoints

## Key Patterns

### Status Codes

- **401 Unauthorized**: Invalid/expired/missing tokens
- **403 Forbidden**: Valid token but insufficient permissions

### Helper Usage

```typescript
const oidcHelpers = new OIDCHelpers();
const token = await oidcHelpers.tokenManager.getValidToken(TEST_USERS.ADMIN);
const result = await oidcHelpers.makeAuthenticatedRequest(request, 'GET', '/api/apps', token.token);
```

## Troubleshooting

- **Port conflicts**: Ensure ports 8001, 8003, 8063 are available
- **Container logs**: `docker compose -f docker-compose.base.yaml -f docker-compose.oidc-test.yaml logs [service]`
- **Environment**: Tests auto-detect CI vs local environment

## Configuration

All test configuration is centralized in `test-config.ts`:

- Nebraska base URL
- Keycloak settings
- Database connection
