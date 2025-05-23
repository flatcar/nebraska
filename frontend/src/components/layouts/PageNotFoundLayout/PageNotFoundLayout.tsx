import { Grid, Typography } from '@mui/material';
import { Trans, useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';

import pageNotFoundGraphic from './404.svg';

export default function PageNotFoundLayout() {
  const { t } = useTranslation('missing');

  return (
    <Grid
      container
      spacing={0}
      direction="column"
      alignItems="center"
      justifyContent="center"
      style={{ minHeight: '100vh', textAlign: 'center' }}
    >
      <Grid size={6}>
        <img src={pageNotFoundGraphic} alt="page not found 404" style={{ maxWidth: '100%' }} />
        <Typography variant="h1" style={{ fontSize: '1.875rem' }}>
          {t('missing|error_page_not_found')}
        </Typography>
        <Typography variant="h2">
          <Trans t={t} ns="missing" i18nKey="nav_head_back_home">
            Head back <Link to="/">home</Link>.
          </Trans>
        </Typography>
      </Grid>
    </Grid>
  );
}
