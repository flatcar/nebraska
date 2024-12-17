import { deDE, enUS, esES, hiIN, ptPT } from '@mui/material/locale';
import {
  adaptV4Theme,
  createTheme,
  StyledEngineProvider,
  Theme,
  ThemeProvider,
} from '@mui/material/styles';
import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';

declare module '@mui/styles/defaultTheme' {
  // eslint-disable-next-line @typescript-eslint/no-empty-interface
  interface DefaultTheme extends Theme {}
}

function getLocale(locale: string): typeof enUS {
  const LOCALES = {
    en: enUS,
    pt: ptPT,
    es: esES,
    de: deDE,
    ta: enUS, // @todo: material ui needs a translation for this.
    hi: hiIN,
  };
  type LocalesType = 'en' | 'pt' | 'es' | 'ta' | 'de' | 'hi';
  return locale in LOCALES ? LOCALES[locale as LocalesType] : LOCALES['en'];
}

/** Like a ThemeProvider but uses reacti18next for the language selection
 *  Because Material UI is localized as well.
 */
const ThemeProviderNexti18n: React.FC<React.PropsWithChildren<{ theme: Theme }>> = props => {
  const { i18n } = useTranslation();
  const [lang, setLang] = useState(i18n.language);

  function changeLang(lng: string) {
    if (lng) {
      document.documentElement.lang = lng;
      document.body.dir = i18n.dir();
      setLang(lng);
    }
  }

  useEffect(() => {
    i18n.on('languageChanged', changeLang);
    return () => {
      i18n.off('languageChanged', changeLang);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const theme = createTheme(adaptV4Theme(props.theme, getLocale(lang)));

  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>{props.children}</ThemeProvider>
    </StyledEngineProvider>
  );
};

export default ThemeProviderNexti18n;
