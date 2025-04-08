import type { TestRunnerConfig } from '@storybook/test-runner';
import { expect } from 'vitest';

const config: TestRunnerConfig = {
  // Hook that is executed before the test runner starts running tests
  setup() {
    // Add your configuration here.
  },
  /* Hook to execute before a story is initially visited before being rendered in the browser.
   * The page argument is the Playwright's page object for the story.
   * The context argument is a Storybook object containing the story's id, title, and name.
   */
  // async preVisit(page, context) {
  // },
  /* Hook to execute after a story is visited and fully rendered.
   * The page argument is the Playwright's page object for the story
   * The context argument is a Storybook object containing the story's id, title, and name.
   */
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  async postVisit(page, _) {
    const elementHandler = await page.$('#storybook-root');
    if (elementHandler) {
      const innerHTML = await elementHandler.innerHTML();
      expect(innerHTML).toMatchSnapshot();
    } else {
      throw new Error("The '#storybook-root' element was not found. elementHandler is null!");
    }
  },
};

export default config;
