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

const BRAND_TEAL = '#09BAC8';
const BRAND_GOLD = '#FEBA00';
const PRIMARY_TEAL = '#0B7C85';

const lightTheme = createTheme({
  components: {
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: '#F4F6F8',
          backgroundImage:
            'radial-gradient(ellipse 120% 80% at 0% -20%, rgba(9, 186, 200, 0.08), transparent 50%), radial-gradient(ellipse 80% 60% at 100% 0%, rgba(254, 186, 0, 0.06), transparent 45%)',
          minHeight: '100vh',
        },
      },
    },
    MuiSelect: {
      defaultProps: {
        variant: 'outlined',
      },
    },
    MuiFormControl: {
      defaultProps: {
        variant: 'outlined',
      },
    },
    MuiTextField: {
      defaultProps: {
        variant: 'outlined',
        size: 'small',
      },
    },
    MuiInputLabel: {
      defaultProps: {
        variant: 'outlined',
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          '--AppBar-background': '#fff',
          color: '#1D2939',
          boxShadow: '0 1px 2px rgba(16, 24, 40, 0.06)',
          borderBottom: '3px solid transparent',
          borderImage: `linear-gradient(90deg, ${BRAND_TEAL}, ${BRAND_GOLD}) 1`,
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
        },
        elevation1: {
          border: '1px solid #E4E7EC',
          boxShadow: '0 1px 2px rgba(16, 24, 40, 0.04), 0 4px 12px rgba(16, 24, 40, 0.06)',
          borderRadius: 12,
        },
      },
    },
    MuiButton: {
      defaultProps: {
        disableElevation: true,
      },
      styleOverrides: {
        root: {
          textTransform: 'none',
          fontWeight: 600,
          borderRadius: 8,
        },
        containedPrimary: {
          backgroundColor: PRIMARY_TEAL,
          '&:hover': {
            backgroundColor: '#096870',
          },
        },
        outlined: {
          borderColor: '#D0D5DD',
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          borderRadius: 16,
          border: '1px solid #E4E7EC',
          boxShadow: '0 8px 24px rgba(16, 24, 40, 0.12), 0 2px 6px rgba(16, 24, 40, 0.06)',
        },
      },
    },
    MuiDialogTitle: {
      styleOverrides: {
        root: {
          fontSize: '1.25rem',
          fontWeight: 700,
          letterSpacing: '-0.01em',
          padding: '20px 24px 12px',
          borderBottom: '1px solid #EAECF0',
        },
      },
    },
    MuiDialogContent: {
      styleOverrides: {
        root: {
          padding: '20px 24px !important',
          '&:first-of-type': {
            paddingTop: '20px !important',
          },
          '& .MuiTextField-root + .MuiTextField-root, & .MuiFormControl-root + .MuiFormControl-root, & .MuiTextField-root + .MuiFormControl-root, & .MuiFormControl-root + .MuiTextField-root':
            {
              marginTop: 8,
            },
        },
      },
    },
    MuiDialogActions: {
      styleOverrides: {
        root: {
          padding: '12px 24px 20px',
          borderTop: '1px solid #EAECF0',
          gap: '8px',
        },
      },
    },
    MuiDivider: {
      styleOverrides: {
        root: {
          borderColor: '#EAECF0',
        },
      },
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: '#1D2939',
          borderRadius: 8,
          fontSize: '0.75rem',
        },
      },
    },
    MuiLink: {
      styleOverrides: {
        root: {
          color: PRIMARY_TEAL,
          fontWeight: 600,
        },
      },
    },
  },
  palette: {
    background: {
      default: '#F4F6F8',
      paper: '#FFFFFF',
    },
    primary: {
      contrastText: '#fff',
      main: import.meta.env.VITE_PRIMARY_COLOR ? import.meta.env.VITE_PRIMARY_COLOR : PRIMARY_TEAL,
      light: BRAND_TEAL,
    },
    secondary: {
      main: BRAND_GOLD,
      contrastText: '#1D2939',
    },
    text: {
      primary: '#1D2939',
      secondary: '#5D6B7A',
    },
    success: {
      main: green['800'],
      ...green,
    },
  },
  typography: {
    fontFamily: 'Overpass, Roboto, sans-serif',
    body1: {
      fontSize: '0.875rem',
    },
    h1: {
      fontSize: '1.6rem',
      fontWeight: 800,
      letterSpacing: '-0.01em',
    },
    h2: {
      fontSize: '1.5rem',
      fontWeight: 700,
      letterSpacing: '-0.01em',
    },
    h3: {
      fontSize: '1.3rem',
      fontWeight: 700,
    },
    h4: {
      fontSize: '1.15rem',
      fontWeight: 700,
    },
    subtitle1: {
      fontSize: '0.875rem',
      color: 'rgba(0,0,0,0.6)',
    },
    button: {
      fontWeight: 600,
    },
  },
  shape: {
    borderRadius: 10,
  },
});

const darkTheme = createTheme({
  ...lightTheme,
  components: {
    ...lightTheme.components,
    MuiCssBaseline: {
      styleOverrides: {
        body: {
          backgroundColor: '#0F172A',
          backgroundImage: 'none',
          minHeight: '100vh',
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          '--AppBar-background': '#101828',
          borderBottom: '3px solid transparent',
          borderImage: `linear-gradient(90deg, ${BRAND_TEAL}, ${BRAND_GOLD}) 1`,
        },
      },
    },
  },
  palette: {
    mode: 'dark',
    primary: {
      contrastText: '#fff',
      main: BRAND_TEAL,
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
