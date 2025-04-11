# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Documentation & Process:**
  - Added documentation on how Nebraska releases work, streamlining future release procedures.
  - Updated the Helm chart defaults to support the new container image branding and release the Helm chart version 1.2.0.
- **Testing:**
  - Introduced Playwright e2e tests for frontend workflows, integrated into our CI pipeline.
  
---

### Changed
- **Backend Improvements:**
  - Upgraded core Go dependencies:
    - Bumped Go to version 1.23.
    - Updated authentication libraries such as `github.com/coreos/go-oidc/v3` (to v3.14.1) and `golang.org/x/oauth2` (to v0.29.0).
    - Enhanced concurrency support with `golang.org/x/sync` updated to v0.13.0.
  - Enhanced error handling and output:
    - Fixed error formatting in the main application logic.
  - Adjusted default OIDC scopes: Refresh tokens are now requested only if the access token has expired.
- **Frontend Enhancements:**
  - **UI/UX & Testing:**
    - Upgraded Storybook to version 8 and updated story definitions
        - resulting in changes to snapshot tests.
    - Upgraded from React 16 to 18
    - Upgraded from MUI4 to MUI5
        - Updated input field variants and component APIs to reflect newer defaults.
        - Updated snapshots to align with MUI5 styling changes.
        - Fixed several unit tests impacted by stricter type and style checks after upgrading React and MUI.
    - Increased test coverage
    - Made adjustments to ESLint configurations and TypeScript versions to address compatibility with updated libraries (e.g., React, MUI).
    - Fixed unit tests by ensuring proper theme provider wrapping following stricter MUI5 requirements.
  - **Dependency Updates:**
    - Upgraded multiple npm dependencies:
      - I.e., `@typescript-eslint/eslint-plugin` was bumped from 8.18.1 to 8.25.0 and `@types/node` was updated from 14.18.63 to 22.13.1.
    - Updated frontend-related packages:
      - Storybook now uses webpack 5.
      - The API query construction in `api.ts` now uses native URLSearchParams instead of deprecated packages.
  - **Extracted labels and text from the UI into separate language resources** to make the code cleaner and make translation much easier
- **CI/CD & DevOps:**
  - Updated GitHub Actions workflows:
    - Bumped actions such as `actions/setup-python`, `actions/setup-go`, `actions/setup-node`, and `actions/checkout` to their latest versions.
    - Upgraded Docker actions like `docker/setup-buildx-action` and `docker/setup-qemu-action` for enhanced stability.
  - Modernized CI workflows by integrating playwright.
- **General Dependency Updates:**
  - Numerous dependencies across both backend and frontend have been updated, including:
    - `github/labstack/echo/v4` updated from v4.12.0 to v4.13.3.
    - Multiple bumps to dependencies like `golang.org/x/crypto`, `github.com/stretchr/testify`, `github.com/golangci/golangci-lint`, `github.com/jmoiron/sqlx`, and others.
  - Updated several indirect dependencies (e.g., `path-to-regexp`, `express`, `tough-cookie`, `browserify-sign`, `async`, `url-parse`, etc.) ensuring that all parts of the stack align with current best practices.

---

### Deprecated
- Deprecated the use of the legacy `golint` linter in favor of the actively maintained [revive](https://github.com/mgechev/revive).
- Phasing out the old querystring package in frontend code, replacing it with native `URLSearchParams` functionality.
- Gradually deprecating outdated Formik properties in favor of newer component APIs.

---

### Removed
- Removed deprecated Formik render props and unused translation keys, simplifying the codebase.
- Cleaned up legacy Docker Compose configurations now replaced by the standard `docker compose` command in CI workflows.

---

### Fixed
- **Bug Fixes:**
  - Fixed formatting issues.
  - Resolved localization warnings to declutter development console.
  - Addressed snapshot mismatches in Storybook due to style and configuration changes.
- **Tooling & Integration:**
  - Fixed issues in the code generation tool for server APIs by updating `oapi-codegen` and related dependencies.
  - Adjusted CI workflows to prevent version conflicts, especially for tools like `github/codeql-action` and `actions/checkout`.

---

### Security
- **Dependency & Infrastructure Security:**
  - Upgraded Docker base images
  - Upgraded critical security libraries:
    - Updated `github/coreos/go-oidc/v3` and `golang.org/x/oauth2` to patch known vulnerabilities.
    - Bumped `golang.org/x/crypto` to version 0.31.0 in both backend and updater modules.
  - Increased security in session management:
    - Enforced HttpOnly flag on session cookies to mitigate cross-site scripting (XSS) attacks.
- **CI/CD Pipeline Enhancements:**
  - Updated GitHub Actions and Docker build tools to ensure the latest security patches are applied, including:
    - Latest versions for `actions/setup-python`, `actions/setup-node`, `docker/setup-buildx-action`, and `docker/setup-qemu-action`.
  - Upgraded various npm packages to their secure versions.
- **Vulnerability Mitigations:**
  - Regularly addressed and patched dependencies to avoid potential vulnerabilities
