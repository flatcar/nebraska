# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [4.0.0] - 05/02/2026

### Breaking Changes

- **PostgreSQL 14+ is now a hard requirement:** The `lib/pq` driver has been updated to v1.11.1, which only supports PostgreSQL 14 and newer. Previously, PostgreSQL 17.x was the "tested and default" version, but older versions (e.g., PostgreSQL 13) may have still worked. With this update, **operators running PostgreSQL 13 or older must upgrade their database before upgrading Nebraska**. ([#1300](https://github.com/flatcar/nebraska/pull/1300))

### Added

- **Multi-Step Updates with Floor Packages:** Added support for mandatory intermediate update versions (floor packages) that clients must install before reaching the target version. This enables safe migration paths for breaking changes by ensuring clients update through specific versions in order. Floor packages can be configured per channel with optional reasons and are architecture-specific. ([#1195](https://github.com/flatcar/nebraska/pull/1195))
- **Nebraska backend is able to use OIDC userinfo endpoint:** Some OIDC providers do not return group membership inside the access token. The Nebraska frontend passes this access token via the header `Authorization: Bearer <token>` to the backend which can then (optionally) call the OIDC provider's userinfo endpoint to gather group membership. ([#1279](https://github.com/flatcar/nebraska/pull/1279))

### Changed

- **Package Management UI Improvements:**
  - Replaced POST with idempotent PUT operation for floor package management
  - Package list UI now updates immediately after blacklist changes
  - Channel edit dialog filters out blacklisted packages from selection
  - Floor package selection prevents choosing blacklisted packages with clear visual feedback

### Bugfixes

- Fixed package blacklist changes not appearing in UI immediately after save

## [3.0.0] - 28/11/2025

### Semantic Versioning Correction

**This release corrects a versioning mistake in v2.13.0.** Version 2.13.0 contained breaking changes to OIDC authentication that should have triggered a major version bump per [Semantic Versioning](https://semver.org) principles.

**This release contains functionally identical code to v2.13.0.** The only difference is the corrected version number and this documentation update.

#### What this means for you:

- **If you're currently on v2.13.0:** You have already completed the necessary OIDC migration. Updating to 3.0.0 is optional and only corrects the version label. No additional migration or configuration changes are required.

- **If you're on v2.12.0 or earlier:** Upgrade to v3.0.0 and follow the [OIDC Migration Guide](docs/oidc-migration-guide.md). The migration requirements are identical to those documented for v2.13.0.

- **For new deployments:** Use v3.0.0 (not v2.13.0).

We apologize for this versioning mistake and have updated our release process to prevent similar issues in the future.

### Security

- **OIDC Implementation Refactor - Authorization Code Flow with PKCE** ([nebraska#642](https://github.com/flatcar/nebraska/pull/642))
  - Tokens no longer exposed in server logs or query parameters
  - Frontend handles OIDC flow directly with identity provider using PKCE (Proof Key for Code Exchange)
  - In-memory token storage prevents XSS vulnerabilities
  - Stateless backend architecture eliminates session storage related vulnerabilities

### Breaking Changes

- **OIDC Authentication**: Complete refactor requiring migration (see [OIDC Migration Guide](docs/oidc-migration-guide.md))
  - **Removed configuration options**:
    - `--oidc-client-secret` / `NEBRASKA_OIDC_CLIENT_SECRET` - OIDC now requires public client type
    - `--oidc-valid-redirect-urls` - No longer needed with direct frontend flow
    - `--oidc-session-secret` / `NEBRASKA_OIDC_SESSION_SECRET` - Backend is now stateless
    - `--oidc-session-crypt-key` / `NEBRASKA_OIDC_SESSION_CRYPT_KEY` - No server-side sessions
  - **Removed API endpoints**:
    - `GET /login` - Frontend initiates OIDC flow directly with provider
    - `POST /login/token` - Password grant type no longer supported
    - `GET /login/cb` now returns 501 for OIDC mode (GitHub mode only)
  - **Changed default scopes**: From `openid,offline_access` to `openid,profile,email`
  - **Migration requirements**:
    - OIDC provider must be reconfigured from confidential to public client type
    - CORS must be enabled for Nebraska domain on OIDC provider if it is not hosted under the same domain
    - Recommended: Enable session cookies on OIDC provider for seamless SSO re-authentication
      - Configure SSO session duration to 8-12 hours (idle timeout) and 1-7 days (maximum lifetime) based on your security requirements
      - **Keycloak**: Configure "SSO Session Max" and "SSO Session Idle Timeout" under Realm Settings → Sessions
      - **Auth0**: Configure "Maximum Session Lifetime" and "Idle Session Lifetime" under Tenant Settings → Advanced → Session Expiration
      - NOTE: Many times, these SSO session attributes are already set by default
    - When access tokens get lost after page refresh, the OIDC provider automatically re-authenticates users if SSO session is still active (no password re-entry required)
    - Recommended: Configure OIDC provider access token expiration to 1-8 hours (should be less than the SSO maximum session lifetime)

### Changed

- helm/postgresql: temporarily overwrite PostgreSQL subchart images to the Bitnami Legacy registry (`bitnamilegacy/*`) to restore Helm chart deployments after Bitnami Docker Hub deprecations. This is a short-term workaround only; Bitnami Legacy images are archived and will not receive security updates.
- backend: OIDC authentication refactored to use standard SPA authentication pattern with stateless JWT validation ([nebraska#642](https://github.com/flatcar/nebraska/pull/642))
- frontend: Implements OIDC Authorization Code Flow with PKCE directly, removing backend proxy ([nebraska#642](https://github.com/flatcar/nebraska/pull/642))
- api: Note that `oidcCookieAuth` security scheme in OpenAPI spec was never implemented and should be removed in future cleanup


## [v2.13.0] - 21/10/2025 - INCORRECTLY VERSIONED - USE v3.0.0 INSTEAD

**VERSIONING NOTICE:** This release was published with breaking changes as a minor version, which violates semantic versioning. The functionally identical v3.0.0 release corrects this with the proper major version bump. **For new deployments, use v3.0.0 instead.** Existing v2.13.0 users may upgrade to v3.0.0 to align with correct versioning, but no code or configuration changes are required.

### Security
- **OIDC Implementation Refactor - Authorization Code Flow with PKCE** ([nebraska#642](https://github.com/flatcar/nebraska/pull/642))
  - Tokens no longer exposed in server logs or query parameters
  - Frontend handles OIDC flow directly with identity provider using PKCE (Proof Key for Code Exchange)
  - In-memory token storage prevents XSS vulnerabilities
  - Stateless backend architecture eliminates session storage related vulnerabilities

### Breaking Changes
- **OIDC Authentication**: Complete refactor requiring migration (see [OIDC Migration Guide](docs/oidc-migration-guide.md))
  - **Removed configuration options**:
    - `--oidc-client-secret` / `NEBRASKA_OIDC_CLIENT_SECRET` - OIDC now requires public client type
    - `--oidc-valid-redirect-urls` - No longer needed with direct frontend flow
    - `--oidc-session-secret` / `NEBRASKA_OIDC_SESSION_SECRET` - Backend is now stateless
    - `--oidc-session-crypt-key` / `NEBRASKA_OIDC_SESSION_CRYPT_KEY` - No server-side sessions
  - **Removed API endpoints**:
    - `GET /login` - Frontend initiates OIDC flow directly with provider
    - `POST /login/token` - Password grant type no longer supported
    - `GET /login/cb` now returns 501 for OIDC mode (GitHub mode only)
  - **Changed default scopes**: From `openid,offline_access` to `openid,profile,email`
  - **Migration requirements**:
    - OIDC provider must be reconfigured from confidential to public client type
    - CORS must be enabled for Nebraska domain on OIDC provider if it is not hosted under the same domain
    - Recommended: Enable session cookies on OIDC provider for seamless SSO re-authentication
      - Configure SSO session duration to 8-12 hours (idle timeout) and 1-7 days (maximum lifetime) based on your security requirements
      - **Keycloak**: Configure "SSO Session Max" and "SSO Session Idle Timeout" under Realm Settings → Sessions
      - **Auth0**: Configure "Maximum Session Lifetime" and "Idle Session Lifetime" under Tenant Settings → Advanced → Session Expiration
      - NOTE: Many times, these SSO session attributes are already set by default
    - When access tokens get lost after page refresh, the OIDC provider automatically re-authenticates users if SSO session is still active (no password re-entry required)
    - Recommended: Configure OIDC provider access token expiration to 1-8 hours (should be less than the SSO maximum session lifetime)

### Changed

- helm/postgresql: temporarily overwrite PostgreSQL subchart images to the Bitnami Legacy registry (`bitnamilegacy/*`) to restore Helm chart deployments after Bitnami Docker Hub deprecations. This is a short-term workaround only; Bitnami Legacy images are archived and will not receive security updates.
- backend: OIDC authentication refactored to use standard SPA authentication pattern with stateless JWT validation ([nebraska#642](https://github.com/flatcar/nebraska/pull/642))
- frontend: Implements OIDC Authorization Code Flow with PKCE directly, removing backend proxy ([nebraska#642](https://github.com/flatcar/nebraska/pull/642))
- api: Note that `oidcCookieAuth` security scheme in OpenAPI spec was never implemented and should be removed in future cleanup

## [v2.12.0] - 28/08/2025

### Security
- **form-data@4.0.3 → v4.0.4**
  - Fixes CVE-2025-7783 (GHSA-fjxv-7rqg-78g4): Critical vulnerability (CVSS 9.4) where form-data uses Math.random() for selecting multipart/form-data boundary values. This predictable randomness could allow attackers to inject additional parameters into requests (HTTP Parameter Pollution), potentially making arbitrary requests to internal systems. Affected versions: <2.5.4, 3.0.0-3.0.3, 4.0.0-4.0.3. Fixed in: 2.5.4, 3.0.4, 4.0.4. Updated via npm audit fix. ([#1146](https://github.com/flatcar/nebraska/pull/1146))
- **github.com/go-viper/mapstructure/v2 → v2.3.0**  
  - Fixes GHSA-fv92-fjc5-jj9h: Prevents sensitive information leakage in error messages during type conversion failures (https://github.com/flatcar/nebraska/pull/1099)

### Breaking Changes

- Postgresql 17.x is now the tested and default version. For existing Kubernetes deployment, you might need to run a manual intervention (see: [charts/nebraska/README.md](https://github.com/flatcar/nebraska/blob/main/charts/nebraska/README.md#upgrade-postgresql))([nebraska#1088](https://github.com/flatcar/nebraska/pull/1088))

### Added

- helm: add ability to specify extra annotations and labels for pods, PVCs, ingress, deployments, and other resources ([nebraska#1097](https://github.com/flatcar/nebraska/pull/1097/files))

### Changed

- backend: updated kinvolk references to flatcar ([nebraska#1091](https://github.com/flatcar/nebraska/pull/1091/files))
- backend: migrate from go-bindata to embed ([nebraska#1132](https://github.com/flatcar/nebraska/pull/1132))
- backend: update go to v1.24 ([nebraska#1130](https://github.com/flatcar/nebraska/pull/1130/files))
- updater: update go to v1.24 and remove final kinvolk references ([nebraska#1151](https://github.com/flatcar/nebraska/pull/1151/files))

### Bugfixes

## [v2.11.0] - 17/06/2025

### Security

- **Dependency Audit & Vulnerability Mitigation:**
  - In the pursuite of a clean `npm audit` outcome free of vulnerabilities removed legacy CRA/Webpack dependencies that relied on unmaintained libraries, posing ongoing security risks. [See the Internal / Maintenance Notes](#internal--maintenance-notes) for further details.
- **golang.org/x/net → v0.38.0**
  - Fixes CVE-2025-22870 and CVE-2025-22872 in the HTML tokenizer/parser (https://github.com/flatcar/nebraska/pull/1016)
- **golang.org/x/crypto → v0.35.0**
  - Patches CVE-2025-22869 in SSH server implementations to prevent DoS via untransmitted pending content (https://github.com/flatcar/nebraska/pull/1001)

### Added

- **Tooling Upgrades:**
  - Migrated the build system from Webpack (CRA) to Vite to achieve faster build time and modern native ES module support.
- Add `new_release.md` template based on Flatcar release guidelines (https://github.com/flatcar/nebraska/pull/1002)

### Changed

- **TypeScript Enhancements:**
  - Simplified the `tsconfig.json` file and enabled stricter type-checking settings.
  - Addressed some newly surfaced TypeScript warnings and errors
- Bump Helm chart versions: app → v2.10.0, charts → v1.3.0 (https://github.com/flatcar/nebraska/pull/1012)

### Internal / Maintenance Notes

- The legacy CRA/Webpack setup, although reliable in the early stages, has become increasingly difficult to maintain due to slow update cycles and performance inefficiencies.
- Transitioning to Vite and Vitest was driven by the need for a modern build and testing environment that offers faster development cycles, improved maintainability, and a cleaner security profile.
- Switched from Jest to Vitest to leverage a Vite-native testing environment, enhancing overall developer experience with faster test cycles.
- Upgraded react router to v7 (https://github.com/flatcar/nebraska/pull/1048)
- improved HMR
- Updated and standardized ESLint and Prettier configurations.
- Removed the deprecated `headlamp` dependency.
- Introduced `eslint.config.js` to centralize and simplify linting rules.
- Moved `index.html` from the `public/` folder to the root (`frontend/`) to better align with Vite’s optimal project structure and optimized the file for Vite usage.
- Reformatted and cleaned up End-to-End (E2E) test files to enhance clarity and consistency.
- Add badges to README for CI/status/integration (https://github.com/flatcar/nebraska/pull/993)

### Refactors & Improvements

- **MUI Upgrade:**
  - Minor visual improvements
    by upgrading @mui/material, @mui/system, @mui/icons-material, @mui/utils, @mui/styles, @mui/styled-engine-sc to ^7.0.0 (https://github.com/flatcar/nebraska/pull/1040/files)
- **Icon Management:**
  - Refactored the icon builder to utilize the new Iconify API for improved performance and maintainability.
- **General Codebase Improvements:**
  - Optimized import arrangements and removed unused parameters across multiple functions.
  - Enhanced the `package.json` scripts for development, testing, and build processes.
  - Updated various minor and major dependencies to maintain compatibility and stability.

---

## [v2.10.0] - 15/04/2025

### Security

- **Dependency & Infrastructure Security:**
  - Upgraded Docker base images.
  - Updated critical security libraries:
    - Upgraded `github/coreos/go-oidc/v3` and `golang.org/x/oauth2` to patch known vulnerabilities.
    - Bumped `golang.org/x/crypto` to version 0.31.0.
  - Increased session management security by enforcing the HttpOnly flag on session cookies.
- **CI/CD Pipeline Enhancements:**
  - Updated GitHub Actions, Docker build tools, and various npm packages to incorporate the latest security patches.
- **Vulnerability Mitigations:**
  - Patched dependencies to mitigate potential vulnerabilities.

### Added

- **Documentation & Process:**
  - Updated the Helm chart defaults to support new container image branding and released Helm chart version 1.2.0.

### Changed

- **Backend Improvements:**
  - Upgraded core Go dependencies:
    - Bumped Go to version 1.23.
    - Updated authentication libraries (`github/coreos/go-oidc/v3` to v3.14.1 and `golang.org/x/oauth2` to v0.29.0).
    - Enhanced concurrency support with `golang.org/x/sync` updated to v0.13.0.
  - Adjusted default OIDC scopes so that refresh tokens are requested only if the access token has expired.
- **Frontend Enhancements:**
  - **UI/UX Updates:**
    - Upgraded from React 16 to 18.
    - Upgraded from MUI4 to MUI5:
      - Updated input field variants, component APIs, and snapshots to reflect styling changes.
      - Fixed several unit tests affected by stricter type and style checks.
  - **Dependency Updates & Refactoring:**
    - Upgraded multiple npm dependencies (e.g., `@typescript-eslint/eslint-plugin` from 8.18.1 to 8.25.0 and `@types/node` from 14.18.63 to 22.13.1).
    - Updated the API query construction in `api.ts` to use native URLSearchParams, replacing deprecated packages.
    - Extracted UI labels and text into separate language resources to simplify translation.
- **CI/CD & DevOps:**
  - Modernized workflows:
    - Bumped GitHub Actions (e.g., `actions/setup-python`, `actions/setup-go`, `actions/setup-node`, `actions/checkout`) and Docker actions (`docker/setup-buildx-action`, `docker/setup-qemu-action`) to their latest versions.
- **General Dependency Updates:**
  - Updated various backend and frontend dependencies:
    - For example, `github/labstack/echo/v4` was updated from v4.12.0 to v4.13.3.
    - Updated other dependencies such as `github.com/stretchr/testify`, `github.com/golangci/golangci-lint`, `github.com/jmoiron/sqlx`, and several indirect dependencies (e.g., `path-to-regexp`, `express`, `tough-cookie`, `browserify-sign`, `async`, `url-parse`).

### Deprecated

- Deprecated the legacy `golint` linter in favor of the actively maintained [revive](https://github.com/mgechev/revive).
- Phasing out the old querystring package in favor of native `URLSearchParams` in frontend code.
- Gradually deprecating outdated Formik properties in favor of newer component APIs.

### Removed

- Removed deprecated Formik render props and unused translation keys to simplify the codebase.
- Cleaned up legacy Docker Compose configurations, replacing them with the standard `docker compose` command in CI workflows.

### Fixed

- **Tooling & Integration:**
  - Fixed issues with the server API code generation tool by updating `oapi-codegen` and related dependencies.
  - Adjusted CI workflows to prevent version conflicts (e.g., with `github/codeql-action` and `actions/checkout`).

### Internal / Maintenance Notes

- **Testing:**
  - Integrated Playwright end-to-end tests for frontend workflows into the CI pipeline.
  - Increased test coverage.
  - Upgraded Storybook to version 8 with updated story definitions affecting snapshot tests.
- **Tooling Configurations:**
  - Updated ESLint configurations and TypeScript versions for compatibility with newer libraries (e.g., React, MUI).
- **Static Code Analysis & Cleanup:**
  - Fixed formatting issues and resolved localization warnings to declutter the development console.

---
