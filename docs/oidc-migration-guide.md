# OIDC Migration Guide

This guide provides detailed instructions for migrating from the legacy OIDC implementation to the new secure Authorization Code Flow + PKCE architecture introduced in [ADR-001](./architecture-decisions.md#adr-001-oidc-implementation-refactor---authorization-code-flow--pkce-for-spas).

## Overview

The new OIDC implementation eliminates security vulnerabilities by:
- Moving from backend OAuth flow to frontend-direct OIDC communication
- Implementing PKCE (Proof Key for Code Exchange) for enhanced security
- Using in-memory token storage instead of localStorage
- Eliminating token exposure in URLs and server logs

## Prerequisites

- Access to your OIDC provider administration console
- Nebraska deployment configuration access
- Understanding of your current OIDC setup

## Migration Steps

### 1. OIDC Provider Configuration

#### Step 1: Reconfigure Client Type
- Change your OIDC client from **Confidential** to **Public** client type
- Remove the client secret from your OIDC provider configuration
- Enable **Authorization Code Flow** with **PKCE** support

#### Step 2: Configure Redirect URIs
```
https://your-nebraska-domain.com/
https://your-nebraska-domain.com/login/callback
```

#### Step 3: Enable CORS (Critical)
Your OIDC provider must allow CORS requests from your Nebraska frontend domain:
```
Access-Control-Allow-Origin: https://your-nebraska-domain.com
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
```

### Provider-Specific Configuration

#### Keycloak
```bash
# In Keycloak Admin Console:
# 1. Clients → [Your Client] → Settings
# 2. Change "Access Type" from "confidential" to "public"
# 3. Enable "Standard Flow" and "Direct Access Grants"
# 4. Remove "Client Secret" 
# 5. Set "Valid Redirect URIs": https://your-nebraska-domain.com/*
# 6. Set "Web Origins": https://your-nebraska-domain.com
```

#### Auth0
```bash
# In Auth0 Dashboard:
# 1. Applications → [Your App] → Settings
# 2. Change "Application Type" to "Single Page Application"
# 3. Set "Allowed Callback URLs": https://your-nebraska-domain.com/
# 4. Set "Allowed Web Origins": https://your-nebraska-domain.com
# 5. Set "Allowed Origins (CORS)": https://your-nebraska-domain.com
```

#### Okta
```bash
# In Okta Admin Console:
# 1. Applications → [Your App] → General
# 2. Change "Application type" to "Single-page application (SPA)"
# 3. Set "Sign-in redirect URIs": https://your-nebraska-domain.com/
# 4. Set "Trusted Origins" → Add Origin → Type: CORS → Origin: https://your-nebraska-domain.com
```

#### Azure AD / Entra ID
```bash
# In Azure Portal:
# 1. App registrations → [Your App] → Authentication
# 2. Platform configurations → Add "Single-page application"
# 3. Set "Redirect URIs": https://your-nebraska-domain.com/
# 4. Enable "Access tokens" and "ID tokens" 
# 5. API permissions → Ensure proper scopes are granted
```

#### Google Identity
```bash
# In Google Cloud Console:
# 1. APIs & Services → Credentials → [Your OAuth 2.0 Client]
# 2. Application type: "Web application"
# 3. Authorized JavaScript origins: https://your-nebraska-domain.com
# 4. Authorized redirect URIs: https://your-nebraska-domain.com/
```

### 2. Nebraska Configuration Changes

#### Step 1: Remove Deprecated Flags
Remove these flags from your Nebraska deployment configuration:
```bash
# ❌ Remove these OIDC flags (Breaking Changes):
--oidc-client-secret
--oidc-session-secret  
--oidc-session-crypt-key
--oidc-valid-redirect-urls
```

#### Step 2: Update Required Flags
Ensure these flags are properly configured:
```bash
# ✅ Required OIDC configuration:
--auth-mode=oidc
--oidc-client-id=your-public-client-id
--oidc-issuer-url=https://your-oidc-provider.com/realm/your-realm
--oidc-admin-roles=nebraska-admin,admin
--oidc-viewer-roles=nebraska-viewer,member
--oidc-roles-path=roles  # or groups, depending on your token structure
--oidc-scopes=openid,profile,email,roles
--oidc-management-url=https://your-oidc-provider.com/account/
```

#### Step 3: Environment Variables (Alternative)
```bash
# Environment variable configuration:
export NEBRASKA_OIDC_CLIENT_ID=your-public-client-id
export NEBRASKA_AUTH_MODE=oidc
# ... other configuration via environment variables
```

#### Example Deployment Configurations

**Docker Compose:**
```yaml
services:
  nebraska:
    image: nebraska:latest
    environment:
      - NEBRASKA_AUTH_MODE=oidc
      - NEBRASKA_OIDC_CLIENT_ID=your-public-client-id
      - NEBRASKA_OIDC_ISSUER_URL=https://your-oidc-provider.com
      - NEBRASKA_OIDC_ADMIN_ROLES=nebraska-admin
      - NEBRASKA_OIDC_VIEWER_ROLES=nebraska-viewer
      - NEBRASKA_OIDC_SCOPES=openid,profile,email,roles
```

**Kubernetes:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nebraska
spec:
  template:
    spec:
      containers:
      - name: nebraska
        args:
          - --auth-mode=oidc
          - --oidc-client-id=your-public-client-id
          - --oidc-issuer-url=https://your-oidc-provider.com
          - --oidc-admin-roles=nebraska-admin
          - --oidc-viewer-roles=nebraska-viewer
```

### 3. Deployment Verification

#### Step 1: Test OIDC Configuration
1. Access your Nebraska instance: `https://your-nebraska-domain.com`
2. Check that the login redirects to your OIDC provider
3. Complete authentication and verify redirect back to Nebraska
4. Confirm user role-based access works correctly

#### Step 2: Verify Token Validation
```bash
# Test API access with Bearer token:
curl -H "Authorization: Bearer <access-token>" \
     https://your-nebraska-domain.com/config
```

Expected response should include OIDC configuration:
```json
{
  "auth_mode": "oidc",
  "oidc_issuer_url": "https://your-oidc-provider.com",
  "oidc_client_id": "your-public-client-id",
  "oidc_scopes": "openid,profile,email,roles"
}
```

#### Step 3: Check Logs
- ✅ Verify no tokens appear in ingress/server logs
- ✅ Confirm CORS headers are properly set
- ✅ Check for any authentication errors
- ✅ Verify successful JWT token validation

#### Step 4: Test User Flows
1. **Login Flow**: User can authenticate via OIDC provider
2. **Role Authorization**: Admin users can access admin features
3. **Token Refresh**: Expired tokens are handled gracefully
4. **Logout Flow**: Users can logout and are redirected appropriately

### 4. Troubleshooting

#### Common Issues

**Issue: CORS Errors**
```
Access to fetch at 'https://oidc-provider.com' from origin 'https://nebraska.com' has been blocked by CORS policy
```
**Solution**: Ensure your OIDC provider allows CORS from your Nebraska domain

**Issue: Invalid Redirect URI**
```
Error: redirect_uri_mismatch
```
**Solution**: Add your Nebraska domain to allowed redirect URIs in OIDC provider

**Issue: Token Validation Failed**
```
Error: Token verification error
```
**Solution**: Check that roles are properly configured in token claims at the specified `roles-path`

**Issue: User Has No Access**
```
Error: Misconfigured Roles, Can't get access level from access token
```
**Solution**: Verify user has required roles in OIDC provider and roles are included in token

#### Debug Commands

**Check OIDC Provider Metadata:**
```bash
curl https://your-oidc-provider.com/.well-known/openid_configuration
```

**Validate JWT Token:**
```bash
# Use jwt.io or:
echo "your-jwt-token" | cut -d. -f2 | base64 -d | jq .
```

**Test Nebraska Config Endpoint:**
```bash
curl https://your-nebraska-domain.com/config | jq .
```

### 5. Rollback Procedure

If issues occur during migration:

#### Step 1: Temporary Rollback
```bash
# Revert to previous Nebraska version temporarily
# Re-add removed configuration flags with temporary values
--oidc-client-secret=temporary-secret
--oidc-session-secret=temporary-session-secret
```

#### Step 2: Fix OIDC Provider
- Reconfigure your OIDC provider settings
- Ensure CORS is properly enabled
- Verify redirect URIs are correct

#### Step 3: Re-attempt Migration
- Remove temporary flags again
- Test authentication flow thoroughly
- Verify each step of the migration guide

### 6. Security Validation Checklist

After migration, verify:

- [ ] Client secret removed from OIDC provider
- [ ] Client type changed to "Public" or "SPA"
- [ ] CORS properly configured for your domain
- [ ] Redirect URIs include your Nebraska domain
- [ ] No tokens visible in server/ingress logs
- [ ] Role-based authorization working correctly
- [ ] Users can authenticate and access resources
- [ ] Logout functionality redirects properly
- [ ] JWT access tokens are properly validated
- [ ] Token expiration is handled gracefully
- [ ] No sensitive information in browser localStorage

### 7. Performance and Monitoring

#### Metrics to Monitor
- Authentication success/failure rates
- Token validation latency
- CORS preflight request frequency
- User session duration (now shorter due to security)

#### Expected Changes
- **Improved Security**: No token exposure in logs
- **Better Performance**: Stateless backend enables horizontal scaling
- **User Experience**: Users need to re-authenticate on page refresh (security feature)

## Support

For issues during migration:
1. Check the troubleshooting section above
2. Review the [Architecture Decision Record](./architecture-decisions.md#adr-001-oidc-implementation-refactor---authorization-code-flow--pkce-for-spas)
3. Validate OIDC provider configuration
4. Test with a minimal setup first

## References

- [ADR-001: OIDC Implementation Refactor](./architecture-decisions.md#adr-001-oidc-implementation-refactor---authorization-code-flow--pkce-for-spas)
- [RFC 7636 - Proof Key for Code Exchange (PKCE)](https://datatracker.ietf.org/doc/html/rfc7636)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [OAuth 2.0 for Browser-Based Apps](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps)