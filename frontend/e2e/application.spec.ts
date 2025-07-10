import { expect, test } from '@playwright/test';

import { createApplication, deleteApplication, getUniqueTestSuffix } from './helpers';

test.describe('Applications', () => {
  let appName: string;
  let appId: string;
  let cleanupAppName: string;

  test.beforeEach(async ({ page }, testInfo): Promise<void> => {
    const appData = getUniqueTestSuffix(testInfo.title, 'App');
    appName = appData.name;
    appId = appData.id;
    cleanupAppName = appName;

    await page.goto('/');
  });

  test.afterEach(async ({ page }) => {
    await deleteApplication(page, cleanupAppName);

    await expect(page.getByRole('list')).not.toContainText(cleanupAppName);
    await expect(page.getByRole('list')).not.toContainText(appId);
  });

  test('should create new application', async ({ page }) => {
    await createApplication(page, appName, appId);

    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);
  });

  test('should not allow the creation of applications with existing id', async ({ page }) => {
    await createApplication(page, appName, appId);
    await createApplication(page, appName, appId);

    await expect(
      page
        .getByRole('paragraph')
        .filter({ hasText: 'Something went wrong. Check the form or try again' })
    ).toHaveCount(1);
    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('should edit an existing application', async ({ page }) => {
    await createApplication(page, appName, appId);

    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('list')).toContainText(appName);

    const appItem = page.locator('li').filter({ hasText: appName }).first();
    await appItem.waitFor({ state: 'visible' });

    await appItem.getByTestId('more-menu-open-button').click({ force: true });
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    await page.locator('input[name="name"]').click();
    await page.locator('input[name="name"]').fill('');
    await page.locator('input[name="name"]').fill('Application Edit Test');
    await page.locator('input[name="name"]').press('Tab');

    await page.locator('input[name="product_id"]').click();
    await page.locator('input[name="product_id"]').fill('');
    await page.locator('input[name="product_id"]').fill('io.test.app.edit');
    await page.locator('input[name="product_id"]').press('Tab');

    await page.getByLabel('Description').click();
    await page.getByLabel('Description').fill('');
    await page.getByLabel('Description').fill('Edit Test Application');

    await expect(page).toHaveScreenshot('landing.png');

    await page.getByRole('button', { name: 'Update' }).click();

    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('list')).toContainText('Application Edit Test');
    await expect(page.getByRole('list')).toContainText('io.test.app.edit');

    cleanupAppName = 'Application Edit Test';
  });
});
