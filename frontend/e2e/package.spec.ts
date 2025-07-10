import { expect, test } from '@playwright/test';

import {
  addExtraFile,
  createApplication,
  createChannel,
  createPackage,
  deleteApplication,
  getUniqueTestSuffix,
  TIMEOUTS,
} from './helpers';

test.describe('Packages', () => {
  let appName: string;
  let appId: string;
  test.beforeEach(async ({ page }, testInfo) => {
    // Use the same deterministic naming as other tests
    const appData = getUniqueTestSuffix(testInfo.title, 'Pkg');
    appName = appData.name;
    appId = appData.id;

    await page.goto('/');
    await createApplication(page, appName, appId);

    await page.reload();
    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);
  });

  test.afterEach(async ({ page }) => {
    // Clean up after each test instead of afterAll to prevent interference
    try {
      await page.goto('/');
      await deleteApplication(page, appName);
    } catch (error) {
      console.log(`Cleanup failed for ${appName}:`, error);
    }
  });

  test('should open package creation dialog', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: appName }).click();

    await createChannel(page, 'testChannel');

    await page.evaluate(() => window.scrollTo(0, 0));

    await expect(page).toHaveScreenshot('empty-application-before-package-creation.png', {
      fullPage: true,
    });

    await page
      .locator('div')
      .filter({ hasText: /^Packages$/ })
      .first()
      .getByTestId('modal-button')
      .click();

    await page.getByTestId('package-edit-form').getByText('AMD64', { exact: true }).click();
    await expect(page).toHaveScreenshot('arch-dropdown-package-creation.png');
    await page.locator('#menu- div').first().click();
    await page.getByTestId('package-edit-form').getByText('Flatcar', { exact: true }).click();

    await page.locator('#menu- div').first().click();
    await page.getByLabel('Main').locator('div').filter({ hasText: 'URL *' }).click();
    await page
      .getByLabel('URL *')
      .fill('https://update.release.flatcar-linux.net/amd64-usr/4116.0.0/');
    await page.getByLabel('Filename *').click();
    await page.getByLabel('Filename *').fill('flatcar_production_update.gz');
    await page.getByLabel('Description *').click();
    await page.getByLabel('Description *').fill('test');
    await page.getByLabel('Version *').click();
    await page.getByLabel('Version *').fill('dwadad');
    await page.getByLabel('Size *').click();
    await page.getByLabel('Size *').fill('505686594');
    await page.getByLabel('Hash *').click();
    await page.getByLabel('Hash *').fill('x47kqHYwJ9aF+Z8+Ooc+gR4ed6Q=');
    await page.getByLabel('Flatcar Action SHA256 *').click();
    await page
      .getByLabel('Flatcar Action SHA256 *')
      .fill('vrKk75R1A1ru9LfzWHp/C8Ko7NVgReDS7Fz401k9Ms8=');

    await page.locator('input[name="channelsBlacklist"]').scrollIntoViewIfNeeded();
    await page.locator('input[name="channelsBlacklist"]').click({ force: true });
    await expect(page).toHaveScreenshot('blacklist-dropdown-package-creation.png', {
      fullPage: true,
    });
  });

  test('should create package', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: appName }).click();

    await expect(page.getByText('There are no groups for this')).toContainText(
      'There are no groups for this application yet.Groups help you control how you want to distribute updates to a specific set of instances.'
    );

    await createPackage(page, '4116.0.0');
    await page.waitForLoadState('networkidle');
    await page.reload();
    await page.waitForLoadState('networkidle');

    await expect(page.locator('#main')).toContainText('Version: 4116.0.0 (AMD64)');

    await page.keyboard.press('Escape');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForLoadState('networkidle');

    // Click the more menu button for the specific package version
    const packageItem = page.locator('li').filter({ hasText: '4116.0.0' }).first();
    await packageItem.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
    const menuButton = packageItem.getByTestId('more-menu-open-button');
    await menuButton.click({ force: true });
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await expect(page.getByLabel('URL *')).toHaveValue(
      'https://update.release.flatcar-linux.net/amd64-usr/4116.0.0/'
    );
    await expect(page.getByLabel('Filename *')).toHaveValue('flatcar_production_update.gz');
    await expect(page.getByLabel('Description *')).toHaveValue('test');
    await expect(page.getByLabel('Hash *')).toHaveValue('x47kqHYwJ9aF+Z8+Ooc+gR4ed6Q=');
    await expect(page.getByLabel('Flatcar Action SHA256 *')).toHaveValue(
      'vrKk75R1A1ru9LfzWHp/C8Ko7NVgReDS7Fz401k9Ms8='
    );
    await page.getByRole('button', { name: 'Cancel' }).click();

    await addExtraFile(page);
    await page.reload();

    await page.keyboard.press('Escape');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForLoadState('networkidle');

    // Click the more menu button for the specific package version
    const packageItem2 = page.locator('li').filter({ hasText: '4116.0.0' }).first();
    await packageItem2.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
    const menuButton2 = packageItem2.getByTestId('more-menu-open-button');
    await menuButton2.click({ force: true });
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('tab', { name: 'Extra Files' }).click();
    await expect(page.getByTestId('list-item')).toContainText(
      'SHA1 Hash (base64): 3bIy5D+02noyBd0WuWBDm+Fskx8='
    );
    await expect(page.getByTestId('list-item').getByRole('paragraph')).toContainText('oem-qemu.gz');
    await expect(page.getByTestId('list-item')).toContainText(
      'SHA256 Hash (hex): 505e74c6c3228619c5c785b023ca70d60275c6a0d438a70d3f1fca9af5f26c3a'
    );
  });

  test('should delete package', async ({ page }) => {
    await page.goto('/');

    const navigationPromise = page.waitForLoadState('domcontentloaded');
    await page.getByRole('link', { name: appName }).click();
    await navigationPromise;

    await expect(page.getByTestId('empty').first()).toContainText(
      'There are no groups for this application yet.Groups help you control how you want to distribute updates to a specific set of instances.'
    );

    await createPackage(page, '4117.0.0');
    await page.reload();

    // Ensure our test package was created
    await expect(page.locator('#main')).toContainText('Version: 4117.0.0 (AMD64)');

    await page.keyboard.press('Escape');
    await page.waitForLoadState('domcontentloaded');
    await page.waitForLoadState('networkidle');

    // Click the more menu button for the specific package version to delete
    const packageToDelete = page.locator('li').filter({ hasText: '4117.0.0' }).first();
    await packageToDelete.waitFor({ state: 'visible' });

    // Setup dialog handler before clicking the menu
    page.once('dialog', dialog => {
      dialog.accept().catch(() => {});
    });

    const deleteMenuButton = packageToDelete.getByTestId('more-menu-open-button');
    await deleteMenuButton.click({ force: true });

    // Click delete menu item
    const deleteMenuItem = page.getByRole('menuitem', { name: 'Delete' });
    await deleteMenuItem.click({ force: true });

    // Wait for deletion to complete and reload
    await page.waitForLoadState('networkidle');
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Verify package is deleted - check that the specific version is gone
    await expect(page.locator('#main')).not.toContainText('Version: 4117.0.0 (AMD64)');
  });

  test('should test become searchable', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: appName }).click();
    await page.waitForLoadState('networkidle');

    await createPackage(page, '4117.0.0');
    await createPackage(page, '5439.0.0');
    await createPackage(page, '87.194.0');
    await page.reload();

    // Click the channel modal button with force to overcome any overlay
    await page.getByTestId('modal-button').nth(1).click({ force: true });

    // Click on the package picker and start typing
    const packagePicker = page.getByPlaceholder('Pick a package');
    await packagePicker.click();

    // Wait for the package picker dialog to fully load
    await page.waitForSelector(
      '[role="dialog"][aria-labelledby*="Choose a package"], [role="dialog"] h2:has-text("Choose a package")',
      { timeout: 10000 }
    );

    const searchInput = page.getByPlaceholder('Start typing to search a');
    await searchInput.waitFor({ state: 'visible' });

    // First load all packages by filling with empty string and waiting for them to appear
    await searchInput.fill('');
    await page.waitForLoadState('networkidle');

    // Verify some packages are loaded first
    const dialogContent = page.getByRole('dialog', { name: 'Choose a package' });
    await expect(dialogContent).toContainText('4117.0.0');

    // Now search for '5' which should filter to show only 5439.0.0
    await searchInput.fill('5');
    await page.waitForLoadState('networkidle');

    // Verify that searching for '5' shows only the 5439.0.0 package
    await expect(dialogContent).toContainText('5439.0.0');

    // Clear and search for '0' to match all packages
    await searchInput.clear();
    await searchInput.fill('0');
    await page.waitForLoadState('networkidle');

    // Wait for search results to load - should show all 3 packages
    await expect(page.getByRole('dialog', { name: 'Choose a package' })).toContainText('5439.0.0');
    await expect(page.getByRole('dialog', { name: 'Choose a package' })).toContainText('87.194.0');
    await expect(page.getByRole('dialog', { name: 'Choose a package' })).toContainText('4117.0.0');

    await searchInput.clear();
    await searchInput.fill('.19');
    await page.waitForLoadState('networkidle');

    // Should only show the package with .19 in it
    await expect(page.getByRole('dialog', { name: 'Choose a package' })).toContainText('87.194.0');
    await expect(page.getByRole('dialog', { name: 'Choose a package' })).not.toContainText(
      '4117.0.0'
    );
    await expect(page.getByRole('dialog', { name: 'Choose a package' })).not.toContainText(
      '5439.0.0'
    );
  });
});
