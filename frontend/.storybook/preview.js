import React from 'react';
import themesConf from '../src/lib/themes';
import { StylesProvider } from '@mui/styles';
import '../src/i18n/config';
import ThemeProviderNexti18n from '../src/i18n/ThemeProviderNexti18n';
import { StyledEngineProvider } from '@mui/material/styles';

const darkTheme = themesConf['dark'];
const lightTheme = themesConf['light'];

const withThemeProvider = (Story, context) => {
  const backgroundColor = context.globals.backgrounds ? context.globals.backgrounds.value : 'light';
  const theme = backgroundColor !== 'dark' ? lightTheme : darkTheme;

  const ourThemeProvider = (
    <StyledEngineProvider injectFirst>
      <ThemeProviderNexti18n theme={theme}>
        <Story {...context} />
      </ThemeProviderNexti18n>
    </StyledEngineProvider>
  );
  if (process.env.NODE_ENV !== 'test') {
    return ourThemeProvider;
  } else {
    const generateClassName = (rule, styleSheet) =>
      `${styleSheet?.options.classNamePrefix}-${rule.key}`;

    return (
      <StylesProvider generateClassName={generateClassName}>{ourThemeProvider}</StylesProvider>
    );
  }
};
export const decorators = [withThemeProvider];

export const globalTypes = {
  theme: {
    name: 'Theme',
    description: 'Global theme for components',
    defaultValue: 'light',
    toolbar: {
      icon: 'circlehollow',
      items: ['light', 'dark'],
    },
  },
};

export const parameters = {
  backgrounds: {
    values: [
      { name: 'light', value: 'light' },
      { name: 'dark', value: 'dark' },
    ],
  },
  actions: { argTypesRegex: '^on[A-Z].*' },
};
