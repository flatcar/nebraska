import { Page } from '@playwright/test';
import * as crypto from 'crypto';

export function generateSalt(testName: string): string {
  const hash = crypto.createHash('md5').update(testName).digest('base64');
  return hash.replace(/[^a-zA-Z]/g, '').slice(0, 6);
}

export async function createApplication(page: Page, appName: string, appId: string) {
  await page.getByTestId('modal-button').click();
  await page.locator('input[name="name"]').click();
  await page.locator('input[name="name"]').fill(appName);
  await page.locator('input[name="name"]').press('Tab');
  await page.locator('input[name="product_id"]').fill(appId);
  await page.locator('input[name="product_id"]').press('Tab');
  await page.getByLabel('Description').fill('Test Application');
  await page.getByLabel('Description').press('Tab');
  await page.getByLabel('Groups/Channels').click();
  await page.getByRole('option', { name: 'Do not copy' }).click();
  await page.getByRole('button', { name: 'Add' }).click();
}

export async function deleteApplication(page: Page, appName: any) {
  await page.locator('li').filter({ hasText: appName }).getByTestId('more-menu-open-button').click();
  page.once('dialog', dialog => {
    dialog.accept().catch(() => { });
  });
  await page.getByRole('menuitem', { name: 'Delete' }).click();
}

export async function createPackage(page: Page, packageVersion: string) {
  await page.locator('div').filter({ hasText: /^Packages$/ }).first().getByTestId('modal-button').click();

  await page.getByLabel('Main').locator('div').filter({ hasText: 'URL *' }).click();
  await page.getByLabel('URL *').fill('https://update.release.flatcar-linux.net/amd64-usr/4116.0.0/');
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
  await page.getByLabel('Flatcar Action SHA256 *').fill('vrKk75R1A1ru9LfzWHp/C8Ko7NVgReDS7Fz401k9Ms8=');
  await page.getByRole('button', { name: 'Add' }).click();
}

export async function addExtraFile(page: Page) {
  await page.getByTestId('more-menu-open-button').click();
  await page.getByRole('menuitem', { name: 'Edit' }).click();
  await page.getByRole('tab', { name: 'Extra Files' }).click();
  await page.getByLabel('Add File').click();
  await page.getByTestId('list-item').locator('input[name="name"]').click();
  await page.getByTestId('list-item').locator('input[name="name"]').fill('oem-qemu.gz');
  await page.getByTestId('list-item').getByLabel('Size').click();
  await page.getByTestId('list-item').getByLabel('Size').fill('2047');
  await page.getByLabel('SHA1 Hash (base64)').click();
  await page.getByLabel('SHA1 Hash (base64)').fill('3bIy5D+02noyBd0WuWBDm+Fskx8=');
  await page.getByLabel('SHA256 Hash (hex)').click();
  await page.getByLabel('SHA256 Hash (hex)').fill('505e74c6c3228619c5c785b023ca70d60275c6a0d438a70d3f1fca9af5f26c3a');
  await page.getByRole('button', { name: 'Done' }).click();
  await page.getByRole('button', { name: 'Save' }).click();
}

export async function createChannel(page: Page, channelName: string, architecture?: string, packageVersion?: string) {
  await page.getByTestId('modal-button').nth(1).click();
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
    await page.waitForTimeout(500); // Let dropdown load
    await page.waitForSelector('div[role="option"]', { timeout: 4000 });
    await page.locator('div[role="option"]').getByText(packageVersion).click({ force: true, clickCount: 2 });

    await page.getByRole('button', { name: 'Select' }).click();
  }
  await page.getByRole('button', { name: 'Add' }).click();
}

export async function createGroup(page: Page, groupName: string, channel?: string, trackIdentifier?: string) {
  await page.getByTestId('modal-button').first().click();
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
