import { StorybookConfig } from '@storybook/react-webpack5';

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(js|jsx|ts|tsx)'],

  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    '@storybook/preset-create-react-app',
  ],

  framework: '@storybook/react-webpack5',
  staticDirs: ['../public'],

  docs: {},

  typescript: {
    reactDocgen: 'react-docgen-typescript',
  }
};

export default config;
