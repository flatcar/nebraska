import { expect, test } from '@playwright/test';
import { addExtraFile, createApplication, createChannel, createPackage, deleteApplication, generateSalt } from './helpers';

test.describe('Packages', () => {
  let appName: string; let appId: string;

  test.beforeEach(async ({ page }, testInfo) => {
    const appNameSalt = generateSalt(testInfo.title)
    appName = "Test app" + appNameSalt;
    appId = "io.test.app." + appNameSalt;

    await page.goto('http://localhost:8002/');
    await createApplication(page, appName, appId);

    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);
  });

  test.afterAll(async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    await page.goto('http://localhost:8002/');

    await deleteApplication(page, appName);

    await expect(page.getByRole('list')).not.toContainText(appName);
    await expect(page.getByRole('list')).not.toContainText(appId);
  });


  test('should open package creation dialog', async ({ page }) => {
    await page.goto('http://localhost:8002/');
    await page.getByRole('link', { name: appName }).click();

    await createChannel(page, "testChannel");

    await page.evaluate(() => window.scrollTo(0, 0));
    await expect(page).toHaveScreenshot('empty-application-before-package-creation.png', { fullPage: true });

    await page.locator('div').filter({ hasText: /^Packages$/ }).first().getByTestId('modal-button').click();

    await page.getByTestId('package-edit-form').getByText('AMD64', { exact: true }).click();
    await expect(page).toHaveScreenshot('arch-dropdown-package-creation.png');
    await page.locator('#menu- div').first().click();
    await page.getByTestId('package-edit-form').getByText('Flatcar', { exact: true }).click();

    // await expect(page).toHaveScreenshot('type-dropdown-package-creation.png', { fullPage: true });

    await page.locator('#menu- div').first().click();
    await page.getByLabel('Main').locator('div').filter({ hasText: 'URL *' }).click();
    await page.getByLabel('URL *').fill('https://update.release.flatcar-linux.net/amd64-usr/4116.0.0/');
    await page.getByLabel('Filename *').click();
    await page.getByLabel('Filename *').fill('flatcar_production_update.gz');
    await page.getByLabel('Description *').click();
    await page.getByLabel('Description *').fill('test');
    await page.getByLabel('Version *').click();
    await page.getByLabel('Version *').fill("dwadad");
    await page.getByLabel('Size *').click();
    await page.getByLabel('Size *').fill('505686594');
    await page.getByLabel('Hash *').click();
    await page.getByLabel('Hash *').fill('x47kqHYwJ9aF+Z8+Ooc+gR4ed6Q=');
    await page.getByLabel('Flatcar Action SHA256 *').click();
    await page.getByLabel('Flatcar Action SHA256 *').fill('vrKk75R1A1ru9LfzWHp/C8Ko7NVgReDS7Fz401k9Ms8=');

    await page.locator('input[name="channelsBlacklist"]').scrollIntoViewIfNeeded();
    await page.locator('input[name="channelsBlacklist"]').click({ force: true })
    await expect(page).toHaveScreenshot('blacklist-dropdown-package-creation.png', { fullPage: true });
  });

  test('should create package', async ({ page }) => {
    await page.goto('http://localhost:8002/');
    await page.getByRole('link', { name: appName }).click();

    await expect(page.locator('ul')).toContainText('There are no groups for this application yet.Groups help you control how you want to distribute updates to a specific set of instances.');

    await createPackage(page, '4116.0.0');
    await page.reload();

    await expect(page.locator('#main')).toContainText('Version: 4116.0.0 (AMD64)');

    await page.getByTestId('more-menu-open-button').click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await expect(page.getByLabel('URL *')).toHaveValue('https://update.release.flatcar-linux.net/amd64-usr/4116.0.0/');
    await expect(page.getByLabel('Filename *')).toHaveValue('flatcar_production_update.gz');
    await expect(page.getByLabel('Description *')).toHaveValue('test');
    await expect(page.getByLabel('Hash *')).toHaveValue('x47kqHYwJ9aF+Z8+Ooc+gR4ed6Q=');
    await expect(page.getByLabel('Flatcar Action SHA256 *')).toHaveValue('vrKk75R1A1ru9LfzWHp/C8Ko7NVgReDS7Fz401k9Ms8=');
    await page.getByRole('button', { name: 'Cancel' }).click();

    await addExtraFile(page);
    await page.reload();

    await page.getByTestId('more-menu-open-button').click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('tab', { name: 'Extra Files' }).click();
    await expect(page.getByTestId('list-item')).toContainText('SHA1 Hash (base64): 3bIy5D+02noyBd0WuWBDm+Fskx8=');
    await expect(page.getByTestId('list-item').getByRole('paragraph')).toContainText('oem-qemu.gz');
    await expect(page.getByTestId('list-item')).toContainText('SHA256 Hash (hex): 505e74c6c3228619c5c785b023ca70d60275c6a0d438a70d3f1fca9af5f26c3a');
  });

  test('should delete package', async ({ page }) => {
    await page.goto('http://localhost:8002/');
    await page.getByRole('link', { name: appName }).click();

    await expect(page.locator('ul')).toContainText('There are no groups for this application yet.Groups help you control how you want to distribute updates to a specific set of instances.');

    await createPackage(page, '4117.0.0');
    await page.reload();

    await expect(page.locator('#main')).toContainText('Version: 4117.0.0 (AMD64)');

    await page.getByTestId('more-menu-open-button').click();

    page.once('dialog', dialog => {
      dialog.accept().catch(() => { });
    });
    await page.getByRole('menuitem', { name: 'Delete' }).click();

    await page.reload();
    await expect(page.locator('#main')).not.toContainText('Version: 4117.0.0 (AMD64)');
    await expect(page.locator('#main')).toContainText('This application does not have any package yet');
  });

  test('should test become searchable', async ({ page }) => {
    await page.goto('http://localhost:8002/');
    await page.getByRole('link', { name: appName }).click();

    await createPackage(page, '4117.0.0');
    await createPackage(page, '5439.0.0');
    await createPackage(page, '87.194.0');
    await page.reload();

    await page.getByTestId('modal-button').nth(1).click();
    await page.getByPlaceholder('Pick a package').click();
    await page.getByPlaceholder('Start typing to search a').fill('5');

    await expect(page.locator('role=option').filter({ hasText: '5439.0.0' })).toHaveCount(1);
    await expect(page.locator('role=option').filter({ hasText: '4117.0.0' })).toHaveCount(0);
    await expect(page.locator('role=option').filter({ hasText: '87.194.0' })).toHaveCount(0);

    await page.getByPlaceholder('Start typing to search a').click();
    await page.getByPlaceholder('Start typing to search a').fill('0');

    await expect(page.locator('role=option').filter({ hasText: '5439.0.0' })).toHaveCount(1);
    await expect(page.locator('role=option').filter({ hasText: '4117.0.0' })).toHaveCount(1);
    await expect(page.locator('role=option').filter({ hasText: '87.194.0' })).toHaveCount(1);

    await page.getByPlaceholder('Start typing to search a').click();
    await page.getByPlaceholder('Start typing to search a').fill('.19');

    await expect(page.locator('role=option').filter({ hasText: '87.194.0' })).toHaveCount(1);
    await expect(page.locator('role=option').filter({ hasText: '4117.0.0' })).toHaveCount(0);
    await expect(page.locator('role=option').filter({ hasText: '5439.0.0' })).toHaveCount(0);
  })
});

