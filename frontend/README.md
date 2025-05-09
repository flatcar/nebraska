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

# Start development server
npm run dev

# To run vitests
npm test

# Run linter and formatter
npm run lint
npm run format

# Build for production
npm run build

# Run Storybook
npm run storybook

## To update storybook snapshots
npm run build-storybook:ci && npm run serve-storybook:ci
npm run test-storybook:ci -- -u

# Build Storybook
npm run build-storybook

# Generate test coverage report
npm run test:coverage
