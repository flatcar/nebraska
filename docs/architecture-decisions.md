# Architecture Decision Records (ADRs)

This document captures important architectural decisions made for the Nebraska project, including context, options considered, and rationale for each decision.

---

## ADR-001: OIDC Implementation Refactor - Authorization Code Flow + PKCE for SPAs

**Status**: Implemented  
**Date**: 2025-01-10  
**Issue**: [#642 - [SECURITY] OIDC ID token leaking in ingress controller logs](https://github.com/flatcar/nebraska/issues/642)

### Context

**Problem:** OIDC tokens were exposed in server logs via query parameters, creating security vulnerabilities.

**Additional Issues:**
- Deprecated password grant authentication
- localStorage token storage (XSS vulnerable)
- Backend OAuth flow complexity
- Non-standard SPA authentication

### Decision

**Solution: OIDC Authorization Code Flow + PKCE**

**Backend:** Stateless JWT validation only (no OAuth flow, sessions, or client secrets)
**Frontend:** Direct OIDC communication with PKCE and in-memory token storage

### Benefits

**Security:** PKCE protection, in-memory storage, no token exposure in logs, stateless backend
**Architecture:** Standard SPA pattern, improved scalability, simplified codebase, clear separation of concerns

### Migration Impact

**Breaking Changes:** Frontend-direct OIDC flow, public client setup, memory-only tokens
**Compatible:** API authentication, role-based authorization, existing OIDC providers

### Configuration Changes

**Removed Flags:**
- `--oidc-client-secret` (public client, no secret needed)
- `--oidc-session-secret` (stateless backend)
- `--oidc-session-crypt-key` (stateless backend)
- `--oidc-valid-redirect-urls` (provider-side validation)

**Migration Required:** OIDC provider reconfiguration + flag cleanup. See [OIDC Migration Guide](./oidc-migration-guide.md).

### Trade-offs

**Easier:** Security compliance, horizontal scaling, standard SPA patterns, simpler maintenance  
**More difficult:** Page refresh re-authentication, CORS requirements, public client setup

### Current Limitations

**Session Persistence:** Users re-authenticate on page refresh (tokens in memory)  
**Session Duration:** Limited to access token lifetime (15-60 minutes)

### Why Refresh Tokens Are Not Required

For Nebraska's use case as an infrastructure admin tool:

**Usage Pattern:** Administrators typically use Nebraska a few times per month for specific maintenance tasks  
**Session Requirements:** SSO sessions (8-12 hours) exceed typical usage duration  
**User Experience:** SSO provides seamless re-authentication without manual intervention  
**Complexity Trade-off:** Refresh token implementation adds significant complexity for minimal benefit given the usage pattern  

The OIDC provider's SSO session cookies handle re-authentication transparently, making refresh tokens unnecessary for this low-frequency admin tool use case. Users get the same "stay logged in" experience without the additional implementation and security complexity of refresh token rotation, storage, and revocation mechanisms.

### Priority TODOs

**Multi-tab Sync:** BroadcastChannel API for consistent authentication state  
**Update Flatcar Website Documentation:** Update Nebraska authentication documentation on the Flatcar website to reflect the new OIDC implementation and migration guide

### References

- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
- [OAuth 2.0 Security BCP](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-security-topics)
- [OAuth 2.0 for SPAs](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-browser-based-apps)
