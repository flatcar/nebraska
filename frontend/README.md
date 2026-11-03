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

### Prerequisites

- Node.js
- Docker (for the backend database)
- `psql` client (PostgreSQL CLI)

### Install dependencies

```bash
npm install
```

## Development

### Start the backend

```bash
docker run --rm -d --name nebraska-postgres-dev -p 5432:5432 -e POSTGRES_PASSWORD=nebraska postgres && \
    sleep 10 && \
    psql postgres://postgres:nebraska@localhost:5432/postgres -c 'create database nebraska;' && \
    psql postgres://postgres:nebraska@localhost:5432/nebraska -c 'set timezone = "utc";'
make run-backend
```

### Run the development server

```bash
npm run dev
```

## Testing

### Run unit tests (Vitest)

```bash
npm test
```

### Update Storybook snapshots

```bash
npm run build-storybook:ci && npm run serve-storybook:ci
npm run test-storybook:ci -- -u
```

### [E2E Playwright tests](./e2e/README.md)

### Generate test coverage report

```bash
npm run test:coverage
```

## Linting & Formatting

```bash
npm run lint
```

```bash
npm run format
```

## Production Build

```bash
npm run build
```

## Storybook

```bash
# Run Storybook locally
npm run storybook

# Build static Storybook
npm run build-storybook
```
