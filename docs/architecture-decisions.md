# Architecture Decision Records (ADRs)

This document captures important architectural decisions made for the Nebraska project, including context, options considered, and rationale for each decision.

## ADR Format

Each ADR follows this structure:
- **Status**: Proposed, Accepted, Deprecated, or Superseded
- **Context**: What is the issue that we're seeing?
- **Decision**: What is the change that we're proposing?
- **Consequences**: What becomes easier or more difficult to do?

---

## ADR-001: OIDC Token Security - Fragment vs Session-Based Authentication

**Status**: Proposed  
**Date**: 2025-01-10  
**Issue**: [#642 - OIDC ID token leaking in ingress controller logs](https://github.com/flatcar/nebraska/issues/642)

### Context

Nebraska's current OIDC implementation has a security vulnerability where ID tokens are passed as URL query parameters during the authentication callback. This violates RFC 6750 recommendations and creates several security risks:

**Current Implementation Issues:**
- ID tokens appear in ingress controller/proxy logs
- Tokens stored in browser history
- Potential exposure in referrer headers
- Anyone with log access can steal authentication tokens

**Current Architecture:**
The existing implementation is a hybrid approach that combines:
- Server-side session storage (refresh tokens, user state)
- Client-side token storage (ID tokens in localStorage)
- Session-based state management with token-based API authentication

This creates an inconsistent security model where the server maintains authentication state but still exposes tokens to potential logging.

### Options Considered

#### Option 1: Fragment Identifier Approach

**Implementation:**
- Change redirect from `?id_token=xxx` to `#id_token=xxx`
- Fragments are not sent to servers, preventing logging
- Minimal code changes to existing architecture

**Pros:**
- ✅ **Minimal disruption**: Single-line change in redirect logic
- ✅ **Maintains current architecture**: No changes to API authentication
- ✅ **Standards compliance**: Aligns with OAuth2/OIDC SPA patterns
- ✅ **Integration-friendly**: APIs remain token-based for external tools
- ✅ **Performance**: No additional server-side lookups
- ✅ **Scalability**: Stateless token validation works with multiple instances
- ✅ **Modern SPA pattern**: Standard for React/Vue/Angular applications

**Cons:**
- ❌ **Still client-side storage**: Tokens remain in localStorage (XSS vulnerable)
- ❌ **Browser history**: Fragments may still appear in browser history
- ❌ **Token lifecycle**: Client must handle token refresh logic

#### Option 2: Pure Session-Based Approach

**Implementation:**
- Remove ID token exposure entirely
- Frontend authenticates using session cookies only
- Backend validates sessions instead of tokens for API access

**Pros:**
- ✅ **Maximum security**: No token exposure in URLs or client storage
- ✅ **XSS protection**: HttpOnly cookies prevent JavaScript access
- ✅ **Server control**: Can revoke sessions immediately
- ✅ **Consistent model**: Single authentication mechanism
- ✅ **Automatic refresh**: Server handles token lifecycle transparently

**Cons:**
- ❌ **Scaling complexity**: Requires shared session storage for multiple instances
- ❌ **API integration difficulty**: Third-party tools need cookie handling
- ❌ **Performance overhead**: Server-side session lookups on each request
- ❌ **Architecture mismatch**: Traditional web app pattern in modern SPA
- ❌ **Mobile/CLI complexity**: Session management challenging for non-browser clients

### Decision

**Recommended: Option 1 (Fragment Identifier Approach)**

**Rationale:**
1. **Immediate security fix** with minimal code changes
2. **Preserves existing architecture** and API compatibility
3. **Aligns with modern standards** and OAuth2 best practices
4. **Maintains scalability** benefits of stateless authentication
5. **Future-proof** for microservices and cloud-native deployments

**Implementation Plan:**
1. Modify `LoginCb` function in `/backend/pkg/auth/oidc.go`
2. Change query parameter to fragment identifier
3. Update frontend token extraction in `/frontend/src/utils/auth.ts`
4. Add security headers to prevent token leakage
5. Update documentation and security guidelines

### Consequences

**What becomes easier:**
- Security compliance with RFC 6750 recommendations
- Preventing token exposure in server logs
- Maintaining current API integration patterns
- Horizontal scaling without shared session storage

**What becomes more difficult:**
- Client-side token management remains complex
- XSS vulnerabilities still exist (though reduced)
- Token lifecycle management stays with the frontend

### Future Considerations

- Consider implementing token binding or proof-of-possession tokens
- Evaluate Content Security Policy (CSP) headers to mitigate XSS risks
- Monitor for additional security improvements in OAuth2/OIDC specifications
- Consider migration to OAuth2 PKCE flow for enhanced security

### References

- [RFC 6750 - OAuth2 Bearer Token Usage](https://datatracker.ietf.org/doc/html/rfc6750#section-2.3)
- [OAuth2 Security Best Current Practice](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)

---

## Template for Future ADRs

```markdown
## ADR-XXX: [Title]

**Status**: [Proposed/Accepted/Deprecated/Superseded]  
**Date**: YYYY-MM-DD  
**Issue**: [Link to relevant issue if applicable]

### Context
[Describe the issue and why it needs to be addressed]

### Options Considered
[List the options with pros and cons]

### Decision
[State the decision and rationale]

### Consequences
[What becomes easier or more difficult]
```