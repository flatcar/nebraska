# Nebraska

**Nebraska** is an update manager for [Flatcar Container Linux](https://www.flatcar.org/), built with **React**, **Vite**, and **MUI** (Material UI).

## Features

- ⚛️ React + Vite
- 🎨 MUI (Material UI) for UI components
- 🌐 i18n with i18next
- 🧪 Testing with Vitest
- 🧼 Linting & formatting via ESLint + Prettier
- 📖 Storybook for isolated UI development

## Getting Started

```bash
# Install dependencies
npm install

# Development
## Start the backend
```bash
docker run --rm -d --name nebraska-postgres-dev -p 5432:5432 -e POSTGRES_PASSWORD=nebraska postgres && \
    sleep 10 && \
    psql postgres://postgres:nebraska@localhost:5432/postgres -c 'create database nebraska;' && \
    psql postgres://postgres:nebraska@localhost:5432/nebraska -c 'set timezone = "utc";'
make run-backend
```
## Run development server
npm run dev

# Run tests
## To run vitests
npm test

## To update storybook snapshots
npm run build-storybook:ci && npm run serve-storybook:ci
npm run test-storybook:ci -- -u

## [E2E Playwright tests](./e2e/README.md)

## Generate test coverage report
npm run test:coverage

## Run linter and formatter
npm run lint
npm run format

# Build for production
npm run build

# Run Storybook
npm run storybook

# Build Storybook
npm run build-storybook
