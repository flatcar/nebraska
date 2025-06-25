# OIDC Migration Guide

Quick migration guide for the new secure OIDC implementation with Authorization Code Flow + PKCE.

## What's Changed
- Frontend handles OIDC flow directly (no backend client secret)
- PKCE security instead of localStorage tokens
- Stateless backend with JWT-only validation

## Prerequisites
- OIDC provider admin access
- Nebraska deployment configuration access

## Migration Steps

### 1. OIDC Provider Configuration

**Required Changes:**
1. Change client type: **Confidential** → **Public** (SPA)
2. Remove client secret
3. Set redirect URI: `https://your-nebraska-domain.com/`
4. Enable CORS for your Nebraska domain

**Quick Provider Setup:**

**Keycloak:** Change "Access Type" to "public", set "Web Origins"  
**Auth0:** Change to "Single Page Application", configure CORS origins  
**Okta:** Change to "Single-page application (SPA)", add Trusted Origins  
**Azure AD:** Add "Single-page application" platform, set redirect URIs  
**Google:** Set "Authorized JavaScript origins" and redirect URIs

### 2. Nebraska Configuration

**Required configuration:**
```bash
--oidc-client-id=your-public-client-id
--oidc-issuer-url=https://your-oidc-provider.com
--oidc-admin-roles=nebraska-admin
--oidc-viewer-roles=nebraska-viewer
```

**Optional configuration:**
```bash
--auth-mode=oidc                           # Authentication mode (default: "oidc")
--oidc-roles-path=roles                    # JSON path for roles in token (default: "roles")
--oidc-scopes=openid,profile,email         # OIDC scopes (default: "openid,profile,email")
--oidc-management-url=https://your-idp.com # URL for account management
```

### 3. Verification

**Test login flow:**
1. Access Nebraska → redirects to OIDC provider
2. Authenticate → redirects back to Nebraska
3. Verify role-based access works

**Verify security:**
- No tokens in server logs
- CORS headers present  
- API authentication with Bearer tokens works

### 4. Troubleshooting

**CORS errors:** Enable CORS for your Nebraska domain in OIDC provider  
**Invalid redirect URI:** Add Nebraska domain to allowed redirect URIs  
**Token validation failed:** Check roles configuration in token claims  
**User has no access:** Verify user roles in OIDC provider and token

**Debug commands:**
```bash
# Check provider metadata
curl https://provider.com/.well-known/openid_configuration

# Validate JWT (use jwt.io)
echo "token" | cut -d. -f2 | base64 -d | jq .

# Test config endpoint
curl https://nebraska.com/config | jq .
```

## References

- [ADR-001: OIDC Implementation Refactor](./architecture-decisions.md#adr-001-oidc-implementation-refactor---authorization-code-flow--pkce-for-spas)
- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
- [OAuth 2.0 Security BCP](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)