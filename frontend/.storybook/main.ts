import { type StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(js|jsx|ts|tsx)'],

  addons: ['@storybook/addon-links', '@storybook/addon-docs'],

  framework: '@storybook/react-vite',
  staticDirs: ['../public'],

  docs: {},

  typescript: {
    reactDocgen: 'react-docgen-typescript',
  }
};

export default config;
