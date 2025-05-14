import { green } from '@mui/material/colors';
import { createTheme, Theme } from '@mui/material/styles';
import React from 'react';

const DISABLE_BROWSER_THEME_PREF = true;

declare module '@mui/material/styles' {
  interface Palette {
    titleColor: '#000000';
    lightSilverShade: '#F0F0F0';
    greyShadeColor: '#474747';
    sapphireColor: '#061751';
  }
}

const lightTheme = createTheme({
  components: {
    MuiSelect: {
      defaultProps: {
        variant: 'standard',
      },
    },
    MuiFormControl: {
      defaultProps: {
        variant: 'standard',
      },
    },
    MuiTextField: {
      defaultProps: {
        variant: 'standard',
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: { '--AppBar-background': '#fff' }
      },
    },
  },
  palette: {
    background: {
      default: '#FAFAFA',
    },
    primary: {
      contrastText: '#fff',
      main: import.meta.env.VITE_PRIMARY_COLOR ? import.meta.env.VITE_PRIMARY_COLOR : '#2C98F0',
    },
    success: {
      main: green['800'],
      ...green,
    },
  },
  typography: {
    fontFamily: 'Overpass, sans-serif',
    body1: {
      fontSize: '0.875rem',
    },
    h1: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    h2: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    h3: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    h4: {
      fontSize: '1.875rem',
      fontWeight: 900,
    },
    subtitle1: {
      fontSize: '0.875rem',
      color: 'rgba(0,0,0,0.6)',
    },
  },
  shape: {
    borderRadius: 0,
  },
});

const darkTheme = createTheme({
  ...lightTheme,
  components: {
    MuiAppBar: {
      styleOverrides: {
        root: { '--AppBar-background': '#000' }
      },
    },
  },
  palette: {
    mode: 'dark',
    primary: {
      contrastText: '#fff',
      main: '#000',
    },
  },
});

export interface ThemesConf {
  [themeName: string]: Theme;
}

const themesConf: ThemesConf = {
  light: lightTheme,
  dark: darkTheme,
};

export default themesConf;

export function usePrefersColorScheme() {
  const mql = window.matchMedia('(prefers-color-scheme: dark)');
  const [value, setValue] = React.useState(mql.matches);

  React.useEffect(() => {
    const handler = (x: MediaQueryListEvent | MediaQueryList) => setValue(x.matches);
    mql.addListener(handler);
    return () => mql.removeListener(handler);
  }, []);

  if (DISABLE_BROWSER_THEME_PREF || typeof window.matchMedia !== 'function') {
    return 'light';
  }

  return value;
}

/**
 * Hook gets theme based on user preference, and also OS/Browser preference.
 * @returns 'light' | 'dark' theme name
 */
export function getThemeName(): string {
  if (DISABLE_BROWSER_THEME_PREF || typeof window.matchMedia !== 'function') {
    return 'light';
  }
  const themePreference: string = localStorage.nebraskaThemePreference;
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
  const prefersLight = window.matchMedia('(prefers-color-scheme: light)').matches;

  let themeName = 'light';
  if (themePreference) {
    // A selected theme preference takes precedence.
    themeName = themePreference;
  } else {
    if (prefersLight) {
      themeName = 'light';
    } else if (prefersDark) {
      themeName = 'dark';
    }
  }

  return themeName;
}

export function setTheme(themeName: string) {
  localStorage.nebraskaThemePreference = themeName;
}
