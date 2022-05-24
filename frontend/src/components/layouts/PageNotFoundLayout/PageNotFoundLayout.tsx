import { Grid, Typography } from '@material-ui/core';
import { Trans, useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import pageNotFoundGraphic from './404.svg';

export default function PageNotFoundLayout() {
  const { t } = useTranslation('404');

  return (
    <Grid
      container
      spacing={0}
      direction="column"
      alignItems="center"
      justify="center"
      style={{ minHeight: '100vh', textAlign: 'center' }}
    >
      <Grid item xs={6}>
        <img src={pageNotFoundGraphic} alt="page not found 404" style={{ maxWidth: '100%' }} />
        <Typography variant="h1" style={{ fontSize: '1.875rem' }}>
          {t(`Whoops! This page doesn't exist`)}
        </Typography>
        <Typography variant="h2">
          <Trans t={t}>
            Head back <Link to="/">home</Link>.
          </Trans>
        </Typography>
      </Grid>
    </Grid>
  );
}
