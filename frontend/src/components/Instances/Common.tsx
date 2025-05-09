import { Box, Link, Theme } from '@mui/material';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import { makeStyles } from '@mui/styles';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router-dom';

const useInstanceCountStyles = makeStyles((theme: Theme) => ({
  instancesCount: {
    fontSize: '2rem;',
    fontWeight: 700,
    paddingBottom: '0.5rem',
    color: theme.palette.greyShadeColor,
  },
  instancesLabel: {
    color: theme.palette.text.secondary,
    fontSize: '1rem;',
    paddingTop: '0.5rem',
  },
  instanceLink: {
    fontSize: '1.2rem',
    color: 'rgb(27, 92, 145)',
  },
}));

export function InstanceCountLabel(props: {
  countText: string | number;
  href?: object;
  instanceListView?: boolean;
  loading?: boolean;
}) {
  const classes = useInstanceCountStyles();
  const { countText, href, instanceListView = false } = props;
  const { t } = useTranslation();

  return (
    <Grid container direction="column">
      <Grid>
        <Typography className={classes.instancesLabel}>{t('instances|instances')}</Typography>
      </Grid>
      <Grid>
        <Typography className={classes.instancesCount}>{countText}</Typography>
      </Grid>
      <Grid>
        {!instanceListView && Number(countText) > 0 ? (
          <Box>
            {!props.loading && (
              <Link to={{ ...href }} component={RouterLink} underline="hover">
                <Typography className={classes.instanceLink}>
                  {t('instances|see_all_instances')}
                </Typography>
              </Link>
            )}
          </Box>
        ) : (
          []
        )}
      </Grid>
    </Grid>
  );
}
