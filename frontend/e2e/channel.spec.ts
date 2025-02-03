import { expect, test } from '@playwright/test';
import { createApplication, createChannel, createPackage, deleteApplication, generateSalt } from './helpers';

test.describe('Channels', () => {

  let appName: string; let appId: string;

  test.beforeEach(async ({ page }, testInfo) => {
    const appNameSalt = generateSalt(testInfo.title);
    appName = "Test app" + appNameSalt;
    appId = "io.test.app." + appNameSalt;

    await page.goto('http://localhost:8002/');
    await createApplication(page, appName, appId);

    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);

    await page.getByRole('link', { name: appName }).click();

    await createPackage(page, '4117.0.0');
    await createPackage(page, '5439.0.0');
    await createPackage(page, '87.194.0');

    await page.goto('http://localhost:8002/');
    await page.getByRole('link', { name: appName }).click();
  });

  test.afterEach(async ({ page }) => {
    await page.goto('http://localhost:8002/');

    await deleteApplication(page, appName);

    await expect(page.getByRole('list')).not.toContainText(appName);
    await expect(page.getByRole('list')).not.toContainText(appId);
  });

  test('should create a package', async ({ page }) => {
    await createChannel(page, 'test1', 'AMD64', '5439.0.0');

    await expect(page.locator('#main')).toContainText('AMD64');
    await expect(page.locator('#main')).toContainText('test1');
    await expect(page.locator('#main')).toContainText('5439.0.0');
  });

  test('should create a second package', async ({ page }) => {
    await createChannel(page, 'test2', 'AMD64', '4117.0.0');

    await expect(page.locator('#main')).toContainText('AMD64');
    await expect(page.locator('#main')).toContainText('test2');
    await expect(page.locator('#main')).toContainText('4117.0.0');
  });

  test('should create a third package', async ({ page }) => {
    await createChannel(page, 'test3', 'X86');

    await expect(page.locator('#main')).toContainText('X86');
    await expect(page.locator('#main')).toContainText('test3');
    await expect(page.locator('#main')).toContainText('No package');
  });
});

