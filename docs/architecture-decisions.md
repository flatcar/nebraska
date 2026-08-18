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

---

## ADR-002: OIDC Access Token Audience Validation

**Status**: Implemented
**Supersedes**: part of ADR-001's stateless JWT validation

### Context

ADR-001 moved the backend to stateless JWT validation. That implementation set
go-oidc's `SkipClientIDCheck: true`, so the backend verified only the signature,
issuer and expiry. Any valid, unexpired token from the configured issuer was
accepted, including one the provider issued to a completely different
application in the same realm or tenant.

### Decision

Validate the access token audience: the `aud` claim must contain the value of
`--oidc-audience`, which becomes required in OIDC mode.

This is the standard OAuth resource-server boundary:

- [RFC 9068](https://www.rfc-editor.org/rfc/rfc9068) §4 requires a resource
  server to reject a JWT access token whose `aud` does not identify it.
- [RFC 8725](https://www.rfc-editor.org/rfc/rfc8725) §3.9 requires audience
  validation to prevent token substitution, and §3.12 lists "use different `aud`
  values for different uses of JWTs from the same issuer" as a way to keep
  validation rules for different token kinds mutually exclusive.

### Options considered

**Validate `azp` / `client_id` instead of or alongside `aud`.** Rejected. `azp`
is optional in OIDC Core and is not a resource-server check; Okta uses `cid`
rather than `azp`, and no comparable project (Kubernetes apiserver, Argo CD,
Grafana, Harbor) pivots on it. Restricting which client applications may call the
API is a separate defence-in-depth feature, not the audience boundary.

**Require the audience to differ from `--oidc-client-id`.** Rejected. A provider
with no separate resource server legitimately issues access tokens whose audience
is the client itself. Dex does exactly that, defaulting the access token audience
to the requesting client. Enforcing a difference would leave such deployments
with no valid configuration.

**Warn instead of failing when the audience is unset.** Rejected as the default.
A deployment that silently keeps accepting foreign tokens is the vulnerable
state, and operators rarely revisit warnings. `--oidc-skip-audience-check` gives
a one-flag path to start immediately without changing the identity provider, so
requiring the audience does not block upgrades.

**Rely on audience validation alone to separate access tokens from ID tokens.**
Adopted for the security fix itself, then hardened separately. Audience
validation is a spec-sanctioned way to keep the two kinds mutually exclusive
(RFC 8725 §3.12), and Keycloak's audience mapper keeps the API audience out of
ID tokens by default. Nebraska additionally rejects tokens that positively
identify as another kind — an RFC 9068 `at+jwt` type header is an authoritative
accept, a Keycloak `typ` claim of `ID` is rejected, and providers that emit no
token type marker are unaffected.

### Consequences

**Breaking** for OIDC deployments: Nebraska refuses to start without
`--oidc-audience`, and providers that do not place that value in the access token
will see requests rejected. Keycloak requires an audience protocol mapper; it
ignores the `audience` request parameter. `--auth-mode=noop` and
`--auth-mode=github` are unaffected. See the
[OIDC Migration Guide](./oidc-migration-guide.md) for the staged upgrade path.

`--oidc-audience` now has two roles: the frontend still sends it as the OIDC
`audience` authorization request parameter, and the backend validates it. These
are the request and the result of the same identifier, so a single setting is
correct. Auth0 in particular needs the request parameter, without which it issues
an opaque access token that cannot be validated at all. Providers that do not
implement the parameter must ignore it per RFC 6749 §3.1 and §3.2.

### References

- [RFC 9068 - JWT Profile for OAuth 2.0 Access Tokens](https://www.rfc-editor.org/rfc/rfc9068)
- [RFC 8725 - JSON Web Token Best Current Practices](https://www.rfc-editor.org/rfc/rfc8725)
- [RFC 9700 - Best Current Practice for OAuth 2.0 Security](https://www.rfc-editor.org/rfc/rfc9700)
- [Keycloak audience support](https://www.keycloak.org/docs/latest/server_admin/#_audience)
