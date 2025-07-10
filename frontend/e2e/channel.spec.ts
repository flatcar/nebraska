import { expect, test } from '@playwright/test';

import {
  createApplication,
  createChannel,
  createPackage,
  deleteApplication,
  getUniqueTestSuffix,
} from './helpers';

test.describe('Channels', () => {
  let appName: string;
  let appId: string;
  let testSalt: string;

  test.beforeEach(async ({ page }, testInfo) => {
    const appData = getUniqueTestSuffix(testInfo.title, 'Channel');
    appName = appData.name;
    appId = appData.id;
    testSalt = appData.salt;

    await page.goto('/');
    await createApplication(page, appName, appId);

    await page.reload();
    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);

    await page.getByRole('link', { name: appName }).click();
  });

  test.afterEach(async ({ page }) => {
    await page.goto('/');

    await deleteApplication(page, appName);

    await expect(page.getByRole('list')).not.toContainText(appName);
    await expect(page.getByRole('list')).not.toContainText(appId);
  });

  test('should create a channel', async ({ page }) => {
    await createPackage(page, '5439.0.0');

    // Ensure we're back to the application page and wait for stability
    await page.waitForLoadState('networkidle');
    await page.keyboard.press('Escape');
    await page.waitForLoadState('domcontentloaded');

    await createChannel(page, 'test1' + testSalt, 'AMD64', '5439.0.0');

    await expect(page.locator('#main')).toContainText('AMD64');
    await expect(page.locator('#main')).toContainText('test1' + testSalt);
    await expect(page.locator('#main')).toContainText('5439.0.0');
  });

  test('should create a second channel', async ({ page }) => {
    await createPackage(page, '4117.0.0');

    // Ensure we're back to the application page and wait for stability
    await page.waitForLoadState('networkidle');
    await page.keyboard.press('Escape');
    await page.waitForLoadState('domcontentloaded');

    await createChannel(page, 'test2' + testSalt, 'AMD64', '4117.0.0');

    await expect(page.locator('#main')).toContainText('AMD64');
    await expect(page.locator('#main')).toContainText('test2' + testSalt);
    await expect(page.locator('#main')).toContainText('4117.0.0');
  });

  test('should create a third channel', async ({ page }) => {
    await createChannel(page, 'test3' + testSalt, 'X86');

    await expect(page.locator('#main')).toContainText('X86');
    await expect(page.locator('#main')).toContainText('test3' + testSalt);
    await expect(page.locator('#main')).toContainText('No package');
  });
});
