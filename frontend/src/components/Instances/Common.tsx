import { Box, Link } from '@mui/material';
import Stack from '@mui/material/Stack';
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
    <Stack direction="column">
      <Box>
        <Typography
          sx={{
            color: theme => theme.palette.text.secondary,
            fontSize: '1rem;',
            paddingTop: '0.5rem',
          }}
        >
          {t('instances|instances')}
        </Typography>
      </Box>
      <Box>
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
      </Box>
      <Box>
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
      </Box>
    </Stack>
  );
}
