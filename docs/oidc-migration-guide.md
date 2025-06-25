# OIDC Migration Guide

Migration guide for Nebraska's secure OIDC implementation with Authorization Code Flow + PKCE.

## What's Changed
- Frontend handles OIDC flow directly (no backend client secret)
- PKCE security for SPA authentication
- Stateless backend with JWT validation

## Migration Steps

### 1. OIDC Provider Configuration

**Required Changes:**
1. Change client type: **Confidential** → **Public** (SPA)
2. Remove client secret
3. Set redirect URI: `https://your-nebraska-domain.com/auth/callback`
4. Enable CORS for your Nebraska domain

**Provider Examples:**
- **Keycloak:** 
  - Change "Access Type" to "public"
  - Add to "Valid Redirect URIs": `https://your-domain.com/auth/callback`
  - Set "Web Origins": `https://your-domain.com`
- **Auth0:** 
  - Change Application Type to "Single Page Application"  
  - Add to "Allowed Callback URLs": `https://your-domain.com/auth/callback`
  - Add to "Allowed Web Origins": `https://your-domain.com`
  - **Important:** Disable "Implicit" grant type, keep only "Authorization Code"
  - Create an API in Auth0 Dashboard → APIs → Create API
    - Set an identifier (e.g., `https://nebraska-api` - doesn't need to be a real URL)
    - Use this identifier as the audience parameter in Nebraska config
- **Okta:** 
  - Change to "SPA" application type
  - Add to "Sign-in redirect URIs": `https://your-domain.com/auth/callback`
  - Add to "Trusted Origins": `https://your-domain.com`
- **Azure AD:**
  - Set Platform to "Single-page application"
  - Add to "Redirect URIs": `https://your-domain.com/auth/callback`  

### 2. Nebraska Configuration

**Required:**
```bash
--oidc-client-id=your-public-client-id
--oidc-issuer-url=https://your-oidc-provider.com
--oidc-admin-roles=nebraska-admin
--oidc-viewer-roles=nebraska-viewer
```

**Optional:**
```bash
--oidc-roles-path=roles                    # JSON path for roles (default: "roles")
--oidc-scopes=openid,profile,email         # OIDC scopes (default: "openid,profile,email")
--oidc-management-url=https://your-idp.com # Account management URL
--oidc-logout-url=https://your-idp.com/logout # Fallback logout URL
--oidc-audience=https://nebraska-api       # Required for Auth0 (use your API identifier)
```

### 3. Verification

1. Access Nebraska → redirects to OIDC provider
2. Authenticate → redirects back to Nebraska
3. Verify role-based access (admin vs viewer)

### 4. Troubleshooting

| Issue | Solution |
|-------|----------|
| CORS errors | Enable CORS for Nebraska domain in OIDC provider |
| Invalid redirect URI | Add `https://your-domain.com/auth/callback` to allowed redirect URIs |
| Token validation failed | Check roles configuration and token claims |
| User has no access | Verify user roles match configured roles |
| JWT decode error (Auth0) | Ensure audience is set and Implicit grant is disabled |

**Debug JWT claims:**
```bash
# Decode JWT payload
echo "token" | cut -d. -f2 | base64 -d | jq .
```

## References

- [ADR-001: OIDC Implementation](./architecture-decisions.md#adr-001-oidc-implementation-refactor)
- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)