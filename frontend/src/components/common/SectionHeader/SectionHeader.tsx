import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import { Box } from '@mui/material';
import Breadcrumbs from '@mui/material/Breadcrumbs';
import Grid from '@mui/material/Grid';
import Link from '@mui/material/Link';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router';

import PageTitle from '../PageTitle/PageTitle';

const PREFIX = 'SectionHeader';

const classes = {
  breadCrumbsItem: `${PREFIX}-breadCrumbsItem`,
};

const StyledBreadcrumbs = styled(Breadcrumbs)(({ theme }) => ({
  [`& .${classes.breadCrumbsItem}`]: {
    '& > a': {
      color: theme.palette.text.secondary,
    },
  },
}));

interface SectionHeaderProps {
  breadcrumbs: {
    label: string;
    path: string | null;
  }[];
  title: string;
}

export default function SectionHeader(props: SectionHeaderProps) {
  const { t } = useTranslation();
  const { breadcrumbs, title } = props;

  return (
    <>
      <PageTitle title={title} />
      <Grid
        container
        alignItems="center"
        justifyContent="flex-start"
        sx={{
          padding: theme => theme.spacing(1),
          flexShrink: 1,
          marginTop: theme => theme.spacing(1),
          marginBottom: theme => theme.spacing(1),
          display: 'inline-block',
        }}
      >
        <Grid>
          <StyledBreadcrumbs
            aria-label={t('common|breadcrumbs_label').toString()}
            separator={<NavigateNextIcon fontSize="small" />}
          >
            {breadcrumbs &&
              breadcrumbs.map(({ path = null, label }, index) => {
                if (path)
                  return (
                    <Box
                      component="span"
                      className={classes.breadCrumbsItem}
                      key={'breadcrumb_' + index}
                    >
                      <Link to={path} component={RouterLink} underline="hover">
                        {label}
                      </Link>
                    </Box>
                  );
                else
                  return (
                    <Typography key={'breadcrumb_' + index} color="textPrimary">
                      {label}
                    </Typography>
                  );
              })}
            {title && <Typography color="textPrimary">{title}</Typography>}
          </StyledBreadcrumbs>
        </Grid>
      </Grid>
    </>
  );
}
