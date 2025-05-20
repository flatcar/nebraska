import { Box, Link } from '@mui/material';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router';

export function InstanceCountLabel(props: {
  countText: string | number;
  href?: object;
  instanceListView?: boolean;
  loading?: boolean;
}) {
  const { countText, href, instanceListView = false } = props;
  const { t } = useTranslation();

  return (
    <Grid container direction="column">
      <Grid>
        <Typography
          sx={{
            color: theme => theme.palette.text.secondary,
            fontSize: '1rem;',
            paddingTop: '0.5rem',
          }}
        >
          {t('instances|instances')}
        </Typography>
      </Grid>
      <Grid>
        <Typography
          sx={{
            fontSize: '2rem;',
            fontWeight: 700,
            paddingBottom: '0.5rem',
            color: theme => theme.palette.greyShadeColor,
          }}
        >
          {countText}
        </Typography>
      </Grid>
      <Grid>
        {!instanceListView && Number(countText) > 0 ? (
          <Box>
            {!props.loading && (
              <Link to={{ ...href }} component={RouterLink} underline="hover">
                <Typography
                  sx={{
                    fontSize: '1.2rem',
                    color: 'rgb(27, 92, 145)',
                  }}
                >
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
