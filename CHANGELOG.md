# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Overview

This release brings significant improvements in security, performance, and developer experience for Nebraska. Key updates include enhancements to the backend authentication and concurrency, major frontend framework upgrades, and CI/CD optimizations. A noumerous minor dependency updates.

### Backend Enhancements
- **Authentication, OAuth2 & OIDC:**  
  - Improved OIDC token handling: Refresh tokens are now only required when an access token expires.
  - Upgraded `go-oidc` (up to v3.14.1) for better identity management.
  - Updated `golang.org/x/oauth2` and `golang.org/x/sync` libraries, ensuring compliance with the latest standards and optimizations.
- **Security & Toolchain:**  
  - Major bump for `golang.org/x/crypto` (up to v0.31.0) across backend and updater.
  - Adjustments in the Go toolchain to align with updated version requirements (transition towards Go 1.23 where applicable).

### Frontend Enhancements
- **UI and Testing:**
  - Migrated to newer Storybook versions and reconfigured component stories with adjustments for webpack 5.
  - Upgraded Material UI (MUI) and adapted component properties (e.g., link styling and variant props) according to the latest migration guides.
  - Increased test coverage with updated snapshots and unit test fixes to accommodate stricter theming in MUI 5.
  - Created e2e testsuit with palywright.
  - Upgraded to React v18.
- **Code Quality:**
  - Improved TypeScript and ESLint configurations to maintain code consistency.
  - Optimized API query construction by replacing deprecated modules with native URLSearchParams.
  - Consolidation of localization resources.

### CI/CD & DevOps Optimizations
- **GitHub Actions & Docker:**  
  - Updated base container images
  - Revamped workflows: Updated actions for Node and Go setup; multiple Docker actions were incrementally upgraded (e.g., setup-buildx and setup-qemu).
  - Streamlined Helm chart tooling with updated `azure/setup-helm` and `helm/kind-action`.
- **Documentation & Release Process:**  
  - Improved release documentation detailing new automated steps, ensuring smoother future releases.

### Minor Dependency Updates and Refinements

A wide range of minor updates and refactorings have been applied across both frontend and backend:
- **Dependency Bumps:** Numerous dependencies were updated for security and performance enhancements.
- **Tooling & Linting:** Transitioned to modern linting tools (e.g., replacing deprecated golint with revive) and refined build scripts.
- **Miscellaneous Fixes:**  
  - Adjustments in code formatting, fixing typescript errors, legacy configuration cleanups, and small refactors and docker-compose settings.
