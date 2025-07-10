import { Page } from '@playwright/test';
import { expect } from '@playwright/test';
import * as crypto from 'crypto';

export const TIMEOUTS = {
  ELEMENT_VISIBLE: 5000,
  NETWORK_IDLE: 10000,
  MODAL_CLOSE: 5000,
  FORM_SUBMISSION: 8000,
} as const;

export async function cleanupTestApplications(page: Page, testPrefix: string = 'Test app') {
  try {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Find all test applications
    const testApps = page.locator('li').filter({ hasText: new RegExp(`${testPrefix}.*`) });
    const count = await testApps.count();

    // Delete each test application
    for (let i = 0; i < count; i++) {
      const appText = await testApps.nth(i).textContent();
      if (appText && appText.includes(testPrefix)) {
        const appNameMatch = appText.match(new RegExp(`(${testPrefix}[^\\s]*)`));
        if (appNameMatch) {
          await deleteApplication(page, appNameMatch[1]);
        }
      }
    }
  } catch (error) {
    // Ignore cleanup errors
    console.log('Error during test cleanup:', error);
  }
}

export function generateSalt(testName: string, workerIndex?: number): string {
  const workerId = workerIndex ?? process.pid % 1000;
  const timestamp = Date.now().toString(36); // Base36 for shorter strings
  const hash = crypto
    .createHash('md5')
    .update(`${testName}-${workerId}-${timestamp}`)
    .digest('base64');
  return hash.replace(/[^a-zA-Z]/g, '').slice(0, 8);
}

export function getUniqueTestSuffix(testTitle: string, prefix: string = 'App') {
  // Create deterministic but unique suffix using a simple hash of the test title
  const hash = testTitle.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  const suffix = `T${hash}`;

  return {
    name: `${prefix} ${suffix}`,
    id: `io.test.${prefix.toLowerCase()}.${suffix.toLowerCase()}`,
    salt: suffix,
  };
}

export async function createApplication(page: Page, appName: string, appId: string) {
  await page.waitForLoadState('networkidle');

  const modalButton = page.getByTestId('modal-button');
  await modalButton.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await modalButton.click();

  const nameInput = page.locator('input[name="name"]');
  await nameInput.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await nameInput.click();
  await nameInput.fill(appName);
  await nameInput.press('Tab');

  const idInput = page.locator('input[name="product_id"]');
  await idInput.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await idInput.fill(appId);
  await idInput.press('Tab');

  const descInput = page.getByLabel('Description');
  await descInput.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await descInput.fill('Test Application');
  await descInput.press('Tab');

  const dropdown = page.getByLabel('Groups/Channels');
  await dropdown.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await dropdown.click();

  const option = page.getByRole('option', { name: 'Do not copy' });
  await option.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await option.click();

  const addButton = page.getByRole('button', { name: 'Add' });
  await addButton.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await addButton.click();

  await page.waitForLoadState('networkidle');
}

export async function deleteApplication(page: Page, appName: string) {
  try {
    const currentUrl = page.url();
    if (!currentUrl.endsWith('/')) {
      await page.goto('/');
    }

    await page.waitForLoadState('networkidle');

    // Close any open modals or menus first
    await page.keyboard.press('Escape');
    await page.keyboard.press('Escape');
    // Wait for any modals/menus to close
    await page.waitForLoadState('domcontentloaded');
    await page.waitForLoadState('networkidle');

    // Additional check for any visible modals and close them
    const modal = page.locator('div[role="presentation"].MuiModal-root');
    if ((await modal.count()) > 0) {
      await page.keyboard.press('Escape');
      // Wait for modal to disappear
      await modal.waitFor({ state: 'hidden', timeout: TIMEOUTS.MODAL_CLOSE }).catch(() => {});
    }

    // Query once and reuse for efficiency
    const appItems = page.locator('li').filter({ hasText: appName });
    const appCount = await appItems.count();

    if (appCount === 0) {
      // App doesn't exist, nothing to delete
      return;
    }

    const appItem = appItems.first();

    await appItem.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });

    page.once('dialog', dialog => {
      dialog.accept().catch(() => {});
    });

    const menuButton = appItem.getByTestId('more-menu-open-button');
    await menuButton.click();

    const deleteMenuItem = page.getByRole('menuitem', { name: 'Delete' });
    await deleteMenuItem.waitFor({ state: 'visible' });

    // Ensure no modals are blocking the click
    const blockingModal = page.locator('div[role="presentation"].MuiDialog-root');
    if ((await blockingModal.count()) > 0) {
      await page.keyboard.press('Escape');
      await blockingModal
        .waitFor({ state: 'hidden', timeout: TIMEOUTS.MODAL_CLOSE })
        .catch(() => {});
    }

    await deleteMenuItem.click({ force: true });

    await page.waitForLoadState('networkidle');

    await page
      .waitForFunction(name => !document.body.innerText.includes(name), appName, {
        timeout: TIMEOUTS.ELEMENT_VISIBLE,
      })
      .catch(() => {
        console.warn(`Could not verify deletion of ${appName}`);
      });
  } catch (error) {
    // Structured error handling with context
    const errorContext = {
      operation: 'deleteApplication',
      appName,
      url: page.url(),
      timestamp: new Date().toISOString(),
    };
    console.error('Application deletion failed:', errorContext, error);

    // Take screenshot for debugging
    await page
      .screenshot({
        path: `./test-results/deletion-error-${Date.now()}.png`,
        fullPage: true,
      })
      .catch(() => {}); // Ignore screenshot errors
  }
}

export async function createPackage(page: Page, packageVersion: string) {
  await page.keyboard.press('Escape');
  await page.waitForLoadState('domcontentloaded');
  await page.waitForLoadState('networkidle');

  await page
    .locator('div')
    .filter({ hasText: /^Packages$/ })
    .first()
    .getByTestId('modal-button')
    .click();

  await page.getByLabel('Main').locator('div').filter({ hasText: 'URL *' }).click();
  await page
    .getByLabel('URL *')
    .fill('https://update.release.flatcar-linux.net/amd64-usr/4116.0.0/');
  await page.getByLabel('Filename *').click();
  await page.getByLabel('Filename *').fill('flatcar_production_update.gz');
  await page.getByLabel('Description *').click();
  await page.getByLabel('Description *').fill('test');
  await page.getByLabel('Version *').click();
  await page.getByLabel('Version *').fill(packageVersion);
  await page.getByLabel('Size *').click();
  await page.getByLabel('Size *').fill('505686594');
  await page.getByLabel('Hash *').click();
  await page.getByLabel('Hash *').fill('x47kqHYwJ9aF+Z8+Ooc+gR4ed6Q=');
  await page.getByLabel('Flatcar Action SHA256 *').click();
  await page
    .getByLabel('Flatcar Action SHA256 *')
    .fill('vrKk75R1A1ru9LfzWHp/C8Ko7NVgReDS7Fz401k9Ms8=');
  await page.getByRole('button', { name: 'Add' }).click();
}

export async function addExtraFile(page: Page) {
  await page.keyboard.press('Escape');
  await page.waitForLoadState('domcontentloaded');

  // Wait for the package list to be stable
  await page.waitForLoadState('networkidle');

  // Click the more menu button for the specific package version 4116.0.0
  const packageItem = page.locator('li').filter({ hasText: '4116.0.0' }).first();
  await packageItem.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });

  const menuButton = packageItem.getByTestId('more-menu-open-button');
  await menuButton.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await menuButton.click();

  // Wait for the menu to appear
  const editMenuItem = page.getByRole('menuitem', { name: 'Edit' });
  await editMenuItem.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await editMenuItem.click();

  // Wait for the edit modal to appear and be ready
  await page.waitForLoadState('networkidle');
  const extraFilesTab = page.getByRole('tab', { name: 'Extra Files' });
  await extraFilesTab.waitFor({ state: 'visible', timeout: TIMEOUTS.ELEMENT_VISIBLE });
  await extraFilesTab.click();
  await page.getByLabel('Add File').click();
  await page.getByTestId('list-item').locator('input[name="name"]').click();
  await page.getByTestId('list-item').locator('input[name="name"]').fill('oem-qemu.gz');
  await page.getByTestId('list-item').getByLabel('Size').click();
  await page.getByTestId('list-item').getByLabel('Size').fill('2047');
  await page.getByLabel('SHA1 Hash (base64)').click();
  await page.getByLabel('SHA1 Hash (base64)').fill('3bIy5D+02noyBd0WuWBDm+Fskx8=');
  await page.getByLabel('SHA256 Hash (hex)').click();
  await page
    .getByLabel('SHA256 Hash (hex)')
    .fill('505e74c6c3228619c5c785b023ca70d60275c6a0d438a70d3f1fca9af5f26c3a');
  await page.getByRole('button', { name: 'Done' }).click();
  await page.getByRole('button', { name: 'Save' }).click();
}

export async function createChannel(
  page: Page,
  channelName: string,
  architecture?: string,
  packageVersion?: string
) {
  const navigationPromise = page.waitForLoadState('domcontentloaded');
  await page.getByTestId('modal-button').nth(1).click();
  await navigationPromise;

  await page.getByTestId('icon-button').click();
  await page.getByTitle('#9900EF').click({ clickCount: 2 });
  await page.getByTitle('#9900EF').press('Escape');

  await expect(page.getByTestId('icon-button').locator('div')).toHaveCSS(
    'background-color',
    'rgb(153, 0, 239)'
  );

  await page.locator('input[name="name"]').click();
  await page.locator('input[name="name"]').fill(channelName);

  if (architecture) {
    // await page.click('div[role = "combobox"] >> text=ARM64');
    await page.getByTestId('channel-edit-form').getByText('AMD64').click();
    await page.getByRole('option', { name: architecture }).click();
  }
  if (packageVersion) {
    await page.getByPlaceholder('Pick a package').click();
    await page.getByPlaceholder('Start typing to search a').fill(packageVersion);
    // Wait for dropdown options to appear instead of hard timeout
    await page.waitForSelector('div[role="option"]', { timeout: TIMEOUTS.ELEMENT_VISIBLE });
    await page
      .locator('div[role="option"]')
      .getByText(packageVersion)
      .click({ force: true, clickCount: 2 });

    await page.getByRole('button', { name: 'Select' }).click();
  }
  await page.getByRole('button', { name: 'Add' }).click();
}

export async function createGroup(
  page: Page,
  groupName: string,
  channel?: string,
  trackIdentifier?: string
) {
  const navigationPromise = page.waitForLoadState('domcontentloaded');
  await page.getByTestId('modal-button').first().click();
  await navigationPromise;

  await page.locator('input[name="name"]').click();
  await page.locator('input[name="name"]').fill(groupName);

  if (channel) {
    await page.getByText('None yet').click();
    await page.getByRole('option', { name: channel }).click();
  }

  if (trackIdentifier) {
    await page.getByLabel('Track (identifier for clients').click();
    await page.getByLabel('Track (identifier for clients').fill(trackIdentifier);
  }

  await page.getByRole('button', { name: 'Add' }).click();
}
