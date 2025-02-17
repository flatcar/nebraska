import i18next from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';
import missing from './locales/en/404.json';
import activity from './locales/en/activity.json';
import applications from './locales/en/applications.json';
import channels from './locales/en/channels.json';
import common from './locales/en/common.json';
import frequent from './locales/en/frequent.json';
import groups from './locales/en/groups.json';
import header from './locales/en/header.json';
import instances from './locales/en/instances.json';
import layouts from './locales/en/layouts.json';
import packages from './locales/en/packages.json';

i18next
  // detect user language https://github.com/i18next/i18next-browser-languageDetector
  .use(LanguageDetector)
  .use(initReactI18next)
  // i18next options: https://www.i18next.com/overview/configuration-options
  .init({
    resources: {
      en: {
        missing,
        activity,
        applications,
        channels,
        common,
        frequent,
        groups,
        header,
        instances,
        layouts,
        packages,
      },
    },
    debug: process.env.NODE_ENV === 'development',
    fallbackLng: 'en',
    supportedLngs: ['en'],
    // nonExplicitSupportedLngs: true,
    interpolation: {
      escapeValue: false, // not needed for react as it escapes by default
      format: function (value, format, lng) {
        // https://www.i18next.com/translation-function/formatting
        if (format === 'number') return new Intl.NumberFormat(lng).format(value);
        if (format === 'date')
          return new Intl.DateTimeFormat(lng, {
            day: '2-digit',
            month: '2-digit',
            year: 'numeric',
          }).format(value);

        return value;
      },
    },
    returnEmptyString: false,
    // https://react.i18next.com/latest/i18next-instance
    // https://www.i18next.com/overview/configuration-options
    react: {
      useSuspense: false, // not needed unless loading from public/locales
      //   bindI18nStore: 'added'
    },
    nsSeparator: '|',
    keySeparator: false,
  });

export default i18next;
