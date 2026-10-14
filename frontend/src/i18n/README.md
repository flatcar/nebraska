# i18n Internationalization / Localization

Nebraska frontend uses [i18next](https://www.i18next.com/) for internationalization. Translation files are organized by namespace in `locales/en/` (`common.json`, `applications.json`, `groups.json`, etc.).

## Using Translations

```typescript
import { useTranslation } from 'react-i18next';

function MyComponent() {
  const { t } = useTranslation();
  return <div>{t('common|confirmation_prompt')}</div>;
}
```

## Extracting Translation Keys

```bash
npm run i18n
```

This scans source files and updates translation JSON files.
