import { expect, test } from '@playwright/test';

import {
  createApplication,
  createChannel,
  createGroup,
  createPackage,
  deleteApplication,
  getUniqueTestSuffix,
} from './helpers';

test.describe('Groups', () => {
  let appName: string;
  let appId: string;
  let testSalt: string;

  test.beforeEach(async ({ page }, testInfo) => {
    const appData = getUniqueTestSuffix(testInfo.title, 'Group');
    appName = appData.name;
    appId = appData.id;
    testSalt = appData.salt;

    await page.goto('/');
    await createApplication(page, appName, appId);

    await page.reload();
    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);

    await page.getByRole('link', { name: appName }).click();

    await createPackage(page, '4117.0.0');
    await page.reload();
    await createChannel(page, 'testChannel' + testSalt, 'AMD64', '4117.0.0');

    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.getByRole('link', { name: appName }).click();
  });

  test.afterEach(async ({ page }) => {
    await page.goto('/');

    await deleteApplication(page, appName);

    await expect(page.getByRole('list')).not.toContainText(appName);
    await expect(page.getByRole('list')).not.toContainText(appId);
  });

  test('should create a group', async ({ page }) => {
    await createGroup(
      page,
      'Test Group 1' + testSalt,
      'testChannel' + testSalt + '(AMD64)',
      'qweqwe123123'
    );

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText(
      'Test Group 1' + testSalt
    );
    await expect(page.getByTestId('list-item')).toContainText('qweqwe123123');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).toContainText('testChannel' + testSalt);
    await expect(page.getByTestId('list-item')).toContainText('4117.0.0 (AMD64)');
  });

  test('should create a group without channel', async ({ page }) => {
    await createGroup(page, 'Test Group 2' + testSalt, undefined, 'qweqwe321321');

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText(
      'Test Group 2' + testSalt
    );
    await expect(page.getByTestId('list-item')).toContainText('qweqwe321321');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).not.toContainText('testChannel');
    await expect(page.getByTestId('list-item')).not.toContainText('4117.0.0 (AMD64)');
  });

  test('should create a group without track identifier', async ({ page }) => {
    await createGroup(page, 'Test Group 3' + testSalt, 'testChannel' + testSalt + '(AMD64)');

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText(
      'Test Group 3' + testSalt
    );
    await expect(page.getByTestId('list-item')).not.toContainText('qweqwe123123');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).toContainText('testChannel' + testSalt);
    await expect(page.getByTestId('list-item')).toContainText('4117.0.0 (AMD64)');
  });

  test('should create a group and update it', async ({ page }) => {
    await createGroup(
      page,
      'Test Group 4' + testSalt,
      'testChannel' + testSalt + '(AMD64)',
      'qweqwe123123'
    );

    await expect(page.getByTestId('list-item').getByRole('link')).toContainText(
      'Test Group 4' + testSalt
    );
    await expect(page.getByTestId('list-item')).toContainText('qweqwe123123');
    await expect(page.getByTestId('list-item')).toContainText('Disabled');
    await expect(page.getByTestId('list-item')).toContainText('testChannel' + testSalt);
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
