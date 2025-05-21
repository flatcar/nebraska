import { expect, test } from '@playwright/test';

import {
  createApplication,
  createChannel,
  createGroup,
  createPackage,
  deleteApplication,
  generateSalt,
} from './helpers';

test.describe('Groups', () => {
  let appName: string;
  let appId: string;

  test.beforeEach(async ({ page }, testInfo) => {
    const appNameSalt = generateSalt(testInfo.title);
    appName = 'Test app' + appNameSalt;
    appId = 'io.test.app.' + appNameSalt;

    await page.goto('/');
    await createApplication(page, appName, appId);

    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);

    await page.getByRole('link', { name: appName }).click();

    await createPackage(page, '4117.0.0');
    await page.reload();
    await createChannel(page, 'testChannel', 'AMD64', '4117.0.0');

    await page.goto('/');
    await page.getByRole('link', { name: appName }).click();
  });

  test.afterEach(async ({ page }) => {
    await page.goto('/');

    await deleteApplication(page, appName);

    await expect(page.getByRole('list')).not.toContainText(appName);
    await expect(page.getByRole('list')).not.toContainText(appId);
  });

  test('should create a group', async ({ page }) => {
    await createGroup(page, 'Test Group 1', 'testChannel(AMD64)', 'qweqwe123123');

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText('Test Group 1');
    await expect(page.getByTestId('list-item')).toContainText('qweqwe123123');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).toContainText('testChannel');
    await expect(page.getByTestId('list-item')).toContainText('4117.0.0 (AMD64)');
  });

  test('should create a group without channel', async ({ page }) => {
    await createGroup(page, 'Test Group 2', undefined, 'qweqwe321321');

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText('Test Group 2');
    await expect(page.getByTestId('list-item')).toContainText('qweqwe321321');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).not.toContainText('testChannel');
    await expect(page.getByTestId('list-item')).not.toContainText('4117.0.0 (AMD64)');
  });

  test('should create a group without track identifier', async ({ page }) => {
    await createGroup(page, 'Test Group 3', 'testChannel(AMD64)');

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText('Test Group 3');
    await expect(page.getByTestId('list-item')).not.toContainText('qweqwe123123');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).toContainText('testChannel');
    await expect(page.getByTestId('list-item')).toContainText('4117.0.0 (AMD64)');
  });

  test('should create a group and update it', async ({ page }) => {
    await createGroup(page, 'Test Group 4', 'testChannel(AMD64)', 'qweqwe123123');

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText('Test Group 4');
    await expect(page.getByTestId('list-item')).toContainText('qweqwe123123');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).toContainText('testChannel');
    await expect(page.getByTestId('list-item')).toContainText('4117.0.0 (AMD64)');

    await page.getByTestId('list-item').getByTestId('more-menu-open-button').click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('tab', { name: 'Policy' }).click();
    await page.getByLabel('Updates enabled').check();
    await page.getByLabel('Safe mode').check();
    await page.getByLabel('Only office hours').check();
    await page.getByPlaceholder('Pick a timezone').click();
    await page.getByText('Africa/Addis_Ababa').click({ clickCount: 2 });
    await page.getByRole('button', { name: 'Select' }).click({ clickCount: 2 });
    await page.getByLabel('hours', { exact: true }).click();
    await page.getByRole('option', { name: 'minutes' }).click();
    await page.getByLabel('days').click();
    await page.getByRole('option', { name: 'days' }).click();

    await page.getByRole('button', { name: 'Save' }).click({ clickCount: 2 });

    await page.getByTestId('list-item').getByRole('link').waitFor({ timeout: 4000 });

    await page.reload();
    await expect(page.getByTestId('list-item')).toContainText('Enabled');
    await expect(page.getByTestId('list-item')).toContainText('Max 1 / 1 minutes');
  });
});
