# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] # CRA to Vite migration

### Security
- **Dependency Audit & Vulnerability Mitigation:**
  - Initially pursued a clean `npm audit` outcome free of vulnerabilities. This goal was re-evaluated as legacy CRA/Webpack dependencies rely on unmaintained libraries, posing ongoing security risks.

### Added
- **Tooling Upgrades:**
  - Migrated the build system from Webpack (CRA) to Vite to achieve faster build time and modern native ES module support.

### Changed
  
- **TypeScript Enhancements:**
  - Simplified the `tsconfig.json` file and enabled stricter type-checking settings.
  - Addressed newly surfaced TypeScript warnings and errors

### Refactors & Improvements
- **Icon Management:**
  - Refactored the icon builder to utilize the new Iconify API for improved performance and maintainability.
  
- **General Codebase Improvements:**
  - Optimized import arrangements and removed unused parameters across multiple functions.
  - Enhanced the `package.json` scripts for development, testing, and build processes.
  - Updated various minor and major dependencies to maintain compatibility and stability.

### Internal / Maintenance Notes
  - The legacy CRA/Webpack setup, although reliable in the early stages, has become increasingly difficult to maintain due to slow update cycles and performance inefficiencies.
  - Transitioning to Vite and Vitest was driven by the need for a modern build and testing environment that offers faster development cycles, improved maintainability, and a cleaner security profile.
  - Switched from Jest to Vitest to leverage a Vite-native testing environment, enhancing overall developer experience with faster test cycles.
  - improved HMR
  - Updated and standardized ESLint and Prettier configurations.
  - Removed the deprecated `headlamp` dependency.
  - Introduced `eslint.config.js` to centralize and simplify linting rules.
  - Moved `index.html` from the `public/` folder to the root (`frontend/`) to better align with Vite’s optimal project structure and optimized the file for Vite usage.
  - Reformatted and cleaned up End-to-End (E2E) test files to enhance clarity and consistency.

---

## [Unreleased] #1

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

