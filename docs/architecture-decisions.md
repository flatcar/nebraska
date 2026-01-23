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

## ADR-002: Multi-Step Updates with Floor Packages

**Status**: Implemented  
**Date**: 2025-10-27  
**PR**: [#1195 - feat: Implement multi-step updates with floor packages](https://github.com/flatcar/nebraska/pull/1195)  
**Discussion**: [Flatcar #1831 - Multi-Step Update Feature Design](https://github.com/flatcar/Flatcar/discussions/1831)

### Context

**Problem:** Nebraska only supported single-step updates, where instances jump directly from their current version to the target. This prevents safe rollout of breaking changes that require intermediate migration steps (e.g., filesystem support changes, partition table migrations).

### Options Considered

1. **Package-level prerequisites**: Each package declares its required predecessor version
2. **Version range specifications**: Complex patterns like `>=3374.2.0`, `!=3450.0.0`, etc.
3. **Channel-level floor packages**: Mandatory checkpoint versions managed per channel

### Decision

**Solution: Channel-level floor packages**

Floor packages are checkpoint versions that ALL clients must install when updating through a channel. They are managed separately from packages in a dedicated `channel_package_floors` junction table.

### Why Floor Packages Over Prerequisites

The package-level prerequisite approach was initially implemented but revealed a critical flaw:

- If package 3602.2.0 requires 3510.2.0 as prerequisite
- Later, 3815.0.0 is released WITHOUT prerequisites if prerquisit is not inherited
- Clients can jump directly to 3815.0.0, bypassing 3510.2.0

This means ALL future packages would need to inherit prerequisites, even for unrelated changes. Floor-based semantics solve this by making checkpoints a channel policy, not a package property. This way we have a central management of update path instead of tracking prerequisits for all later packages individually.

### Architecture

**Database**: New `channel_package_floors` table with channel_id, package_id, floor_reason

**Update Flow**:

- Regular clients: Receive one package at a time, progress sequentially through floors
- Modern syncers (`multi_manifest_ok=true`): Receive floors in batches (up to `NEBRASKA_MAX_FLOORS_PER_RESPONSE`). When more floors remain beyond the limit, response contains only floors (no target). Syncer requests again with highest floor version until all floors are sent, then target is included.
- Legacy syncers: Blocked with `NoUpdate` when floors exist (the syncer itself must be upgraded)

**Safety Rules**:

- Floor/blacklist mutual exclusion (package cannot be both for same channel)
- Channel target cannot be blacklisted
- Architecture must match between floor package and channel

### Configuration

**Environment Variables**:

- `NEBRASKA_MAX_FLOORS_PER_RESPONSE`: Maximum floors per syncer response (default: 5)

**API Endpoints** (see [API spec](../backend/api/spec.yaml) for details):

- `PUT /api/channels/{channelID}/floors/{packageID}` - Set floor (idempotent)
- `DELETE /api/channels/{channelID}/floors/{packageID}` - Remove floor
- `GET /api/channels/{channelID}/floors` - List floors for a channel
- `GET /api/apps/{appIDorProductID}/packages/{packageID}/floor-channels` - List channels where package is a floor

### Trade-offs

**Benefits**:

- Consistent update paths regardless of target version
- Simpler management (channel-level, not per-package)
- No prerequisite inheritance burden on new packages
- Clear separation of policy from package identity

**Limitations**:

- Legacy syncers blocked when floors exist
- Requires careful timing: configure floors BEFORE channel promotion
- Must configure floors for ALL channels (stable, beta, LTS) strategically

### Operational Considerations

**Timing**: Configure floors BEFORE promoting channel to new target. Adding after promotion allows clients to skip floors.

**Cross-channel**: Don't use beta/alpha packages meant as floors for stable (would switch clients to wrong channel).

**Use cases**: Breaking compatibility changes only (e.g., filesystem support), NOT security updates.

### What Was NOT Implemented

From the original design discussion, the following were deferred or rejected:

- Complex version specifications (ranges, patterns, exclusions)
- PostgreSQL semver extension
- Recovery mechanisms
- Canary deployment checkbox for bootloader changes
- Emergency bypass mechanisms

### References

- [Flatcar Issue #1185 - RFE: Multi-stage updates](https://github.com/flatcar/Flatcar/issues/1185)
- [#1195 - feat: Implement multi-step updates with floor packages](https://github.com/flatcar/nebraska/pull/1195)
- [Flatcar #1831 - Multi-Step Update Feature Design](https://github.com/flatcar/Flatcar/discussions/1831)
