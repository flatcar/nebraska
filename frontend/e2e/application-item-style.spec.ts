import { expect, test } from '@playwright/test';

import { createApplication, deleteApplication } from './helpers';

test.describe('Application Item Style', () => {
  let appName: string;
  let appId: string;

  test.beforeEach(async ({ page }): Promise<void> => {
    appName = 'Style Test App';
    appId = 'io.test.style.app.screenshot';

    await page.goto('/');
  });

  test.afterEach(async ({ page }) => {
    await deleteApplication(page, appName);

    await expect(page.getByRole('list')).not.toContainText(appName);
    await expect(page.getByRole('list')).not.toContainText(appId);
  });

  test('should display application item with correct width styling', async ({ page }) => {
    await createApplication(page, appName, appId);

    await expect(page.getByRole('list')).toContainText(appName);
    await expect(page.getByRole('list')).toContainText(appId);

    const appItem = page.locator('li').filter({ hasText: appName });

    await appItem.scrollIntoViewIfNeeded();

    await page.evaluate(() => {
      const items = Array.from(document.querySelectorAll('li'));
      const targetItem = items.find(li => li.textContent?.includes('Test style app'));

      if (targetItem) {
        const rect = targetItem.getBoundingClientRect();
        const scrollTop = window.pageYOffset + rect.top - 100;
        window.scrollTo(0, scrollTop);
      }
    });

    await page.waitForLoadState('networkidle');

    await expect(appItem).toHaveScreenshot('application-item-style.png', {
      clip: {
        x: 0,
        y: -10,
        width: await appItem.boundingBox().then(box => box?.width || 800),
        height: await appItem.boundingBox().then(box => (box?.height || 200) + 20),
      },
    });
  });
});
