## 🏆 Playwright Testing Guide

### 📦 Install Dependencies

```sh
cd frontend
npm install
npx playwright install --with-deps
```

### 🚀 Run Tests

```sh
npx playwright test                                   # Run all tests
npx playwright test file.spec.ts                      # Run specific test file
npx playwright test --ui                              # Run in UI mode
npx playwright test --ui --update-snapshots           # To update snapshots
```

### 🛠 Develop Tests

Create a new test in `e2e/`:

```ts
import { test, expect } from '@playwright/test';

test('example test', async ({ page }) => {
  await page.goto('https://example.com');
  await expect(page).toHaveTitle(/Example/);
});
```

Run the test:

```sh
npx playwright test e2e/example.spec.ts
```

If you use podman, you might need to build the the image first

```sh
podman build --file ../Dockerfile --tag backend_server:local ../ && npx playwright test
```

Use the playwright flags to open the UI

```sh
npx playwright test --fully-parallel --ui
```

Update the snapshots with `-u`

```sh
npx playwright test --fully-parallel -u
```

### 🔍 Debugging

- Use `--debug` to pause execution.
- Open the **HTML report**:
  ```sh
  npx playwright show-report
  ```
- Enable tracing:
  ```sh
  npx playwright test --trace on
  ```

### 🔄 GitHub Actions

1. Push changes to trigger CI tests.
2. Download **playwright-report** from GitHub Actions artifacts.
3. Extract and view the report:
   ```sh
   npx playwright show-report extracted-folder-name
   ```

📖 More: [Playwright Docs](https://playwright.dev)
