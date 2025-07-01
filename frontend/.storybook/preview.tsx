import '../src/i18n/config';

import { StyledEngineProvider } from '@mui/material/styles';
import { createGenerateClassName, StylesProvider } from '@mui/styles';
import type { Preview } from '@storybook/react-vite';
import React from 'react';

import ThemeProviderNexti18n from '../src/i18n/ThemeProviderNexti18n';
import themesConf from '../src/lib/themes';

const darkTheme = themesConf['dark'];
const lightTheme = themesConf['light'];

const isProd = process.env.NODE_ENV === 'production';

// Use stable class names in development & test, but not in production
const generateClassName = !isProd
  ? createGenerateClassName({
    productionPrefix: 'mui',
    disableGlobal: true,
    seed: 'stable', // Ensures stable class names
  })
  : undefined; // In production, let MUI handle class name generation

const withThemeProvider = (Story, context) => {
  const backgroundColor = context.globals.backgrounds?.value || 'light';
  const theme = backgroundColor !== 'dark' ? lightTheme : darkTheme;

  const themeProvider = (
    <StyledEngineProvider injectFirst>
      <ThemeProviderNexti18n theme={theme}>
        <Story {...context} />
      </ThemeProviderNexti18n>
    </StyledEngineProvider>
  );

  return !isProd ? (
    <StylesProvider generateClassName={generateClassName}>{themeProvider}</StylesProvider>
  ) : (
    themeProvider
  );
};

const preview: Preview = {
  decorators: [withThemeProvider],
  globalTypes: {
    theme: {
      name: 'Theme',
      description: 'Global theme for components',
      defaultValue: 'light',
      toolbar: {
        icon: 'circlehollow',
        items: ['light', 'dark'],
      },
    },
  },
  parameters: {
    backgrounds: {
      values: [
        { name: 'light', value: 'light' },
        { name: 'dark', value: 'dark' },
      ],
    },
  },
  tags: ['autodocs']
};

export default preview;
