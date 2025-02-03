import { expect, test } from '@playwright/test';
import { createApplication, deleteApplication, generateSalt } from './helpers';


test.describe('Applications', () => {

  let appName: string; let appId: string;

  test.beforeEach(async ({ page }, testInfo): Promise<void> => {
    const appNameSalt = generateSalt(testInfo.title);
    appName = "Test app" + appNameSalt;
    appId = "io.test.app." + appNameSalt;

    await page.goto('http://localhost:8002/');
  });

  test.afterEach(async ({ page }) => {
    await deleteApplication(page, appName);

    await expect(page.getByRole('list')).not.toContainText(appName);
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

    await expect(page.getByRole('paragraph').filter({ hasText: 'Something went wrong. Check the form or try again' })).toHaveCount(1);
    await page.getByRole('button', { name: 'Cancel' }).click();
  });

  test('should edit an existing application', async ({ page }) => {
    await createApplication(page, appName, appId);

    await page.locator('li').filter({ hasText: appName }).getByTestId('more-menu-open-button').click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.locator('input[name="name"]').click();
    await page.locator('input[name="name"]').fill('Test app modified');
    await page.locator('input[name="name"]').press('Tab');
    await page.locator('input[name="product_id"]').press('ArrowRight');
    await page.locator('input[name="product_id"]').fill('io.test.app.modified');
    await page.locator('input[name="product_id"]').press('Tab');
    await page.getByLabel('Description').press('ArrowRight');
    await page.getByLabel('Description').fill('Test Application modified');

    await expect(page).toHaveScreenshot('landing.png');

    await page.getByRole('button', { name: 'Update' }).click();

    appName = "Test app modified";

    await expect(page.getByRole('list')).toContainText('Test app modified');
    await expect(page.getByRole('list')).toContainText('io.test.app.modified');
  });
});
