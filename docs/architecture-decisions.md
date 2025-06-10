# Architecture Decision Records (ADRs)

This document captures important architectural decisions made for the Nebraska project, including context, options considered, and rationale for each decision.

---

## ADR-001: OIDC Implementation Refactor - Authorization Code Flow + PKCE for SPAs

**Status**: Implemented  
**Date**: 2025-01-10  
**Issue**: [#642 - [SECURITY] OIDC ID token leaking in ingress controller logs](https://github.com/flatcar/nebraska/issues/642)

### Context

**Original Issue Description:**
When using OIDC, the ID token is sent as a query string parameter in GET requests. Many web servers / ingress controllers log the query string.

**Impact:**
A person with simple read access on ingress controllers pods' logs could retrieve an ID token, possibly granting this person more permissions than needed not only on Nebraska but on various other apps also using OIDC.

**Security Vulnerability Evidence:**
```
..."path": "/", "request_query": "id_token=eyJhbG..., "request_length": 3897, "method": "GET"...
```

**Additional Security and Architectural Issues Discovered:**
- Used deprecated password grant type for authentication
- Stored refresh tokens and access tokens in localStorage (XSS vulnerable)  
- Backend handled OAuth flow creating unnecessary complexity
- Hybrid approach mixing session-based and token-based authentication
- Backend stored refresh tokens and managed token lifecycle
- Frontend relied on backend for OAuth flow instead of direct OIDC provider communication
- Non-standard SPA authentication pattern

### Decision

**Implemented: Standard OIDC Authorization Code Flow + PKCE for SPAs**

Completely refactored the OIDC implementation to follow modern SPA security best practices:

#### Backend Changes
- **Removed OAuth flow handling from the backend**: Eliminated `/login` and `/login/cb` endpoints
- **Stateless token validation**: Backend only validates JWT access tokens
- **Removed client secret**: Backend no longer needs confidential client configuration
- **Simplified authentication**: Only extracts roles from JWT access token claims
- **Removed session management**: No more refresh token storage or session handling (which worked inconsistently across the different providers)

#### Frontend Changes  
- **Direct OIDC provider communication**: Frontend handles complete OAuth flow
- **PKCE implementation**: Uses Proof Key for Code Exchange for enhanced security
- **In-memory token storage**: Replaced localStorage with memory storage to prevent XSS
- **Simplified token handling**: Only manages access tokens, no refresh token complexity

### Security Improvements

1. **Eliminated password grant type**: Removed deprecated authentication method
2. **PKCE protection**: Prevents authorization code interception attacks  
3. **In-memory storage**: Tokens no longer persist in localStorage (XSS protection)
4. **No token exposure**: Tokens never appear in URLs or server logs
5. **Stateless backend**: No backend  session storage reduces attack surface
6. **Direct provider communication**: Reduces man-in-the-middle opportunities

### Architecture Benefits

1. **Standard SPA pattern**: Follows RFC 7636 and OAuth2 Security BCP
2. **Improved scalability**: Stateless backend enables horizontal scaling
3. **Simplified codebase**: Removed complex session management logic which worked inconsistently between the auth providers
4. **Better separation of concerns**: Frontend handles authentication flow, backend validates tokens (authorizes)
5. **Future-proof**: Compatible with modern OAuth2/OIDC developments

### Migration Impact

**Breaking Changes:**
- Frontend must handle OIDC flow directly (no backend `/login` endpoint)
- Tokens stored in memory are lost on page refresh (intentional security feature)
- Client configuration requires public client setup (no client secret)

**Backwards Compatibility:**
- API authentication unchanged (still uses Bearer tokens)  
- Same role-based authorization system
- Existing OIDC provider configuration compatible
- Better alignment with existing documentations of auth provider setup

### Consequences

**What becomes easier:**
- Security compliance with OAuth2/OIDC standards
- Horizontal scaling without shared session storage
- Frontend development following modern SPA patterns
- Easier integration with standard OIDC providers
- Maintenance due to simplified architecture

**What becomes more difficult:**
- Users must re-authenticate on page refresh (security vs UX tradeoff)
- OIDC provider must support CORS for direct frontend communication
- Initial setup requires understanding of public client configuration

### Future Considerations

- Evaluate token binding or proof-of-possession tokens for enhanced security
- Monitor OAuth2/OIDC specification updates for additional security features
- Consider implementing refresh tokens if longer session duration is required
- Consider implementing background token validation

### References

- [Authorization Code Flow with Proof Key for Code Exchange (PKCE)](https://auth0.com/docs/get-started/authentication-and-authorization-flow/authorization-code-flow-with-pkce)
- [RFC 7636 - Proof Key for Code Exchange (PKCE)](https://datatracker.ietf.org/doc/html/rfc7636)
- [OAuth 2.0 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [OAuth 2.0 for Browser-Based Apps](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
