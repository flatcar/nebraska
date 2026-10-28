import Tooltip from '@mui/material/Tooltip';
import { useTranslation } from 'react-i18next';

import Label from '../common/Label';

export default function OverrideLabel() {
  const { t } = useTranslation();

  return (
    <Tooltip title={t('groups|override_tooltip')}>
      <span>
        <Label>{t('groups|override')}</Label>
      </span>
    </Tooltip>
  );
}
