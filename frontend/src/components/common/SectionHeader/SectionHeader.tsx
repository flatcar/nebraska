import NavigateNextIcon from '@mui/icons-material/NavigateNext';
import { Box } from '@mui/material';
import Breadcrumbs from '@mui/material/Breadcrumbs';
import Grid from '@mui/material/Grid';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import makeStyles from '@mui/styles/makeStyles';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router-dom';
import PageTitle from '../PageTitle/PageTitle';

const useStyles = makeStyles(theme => ({
  sectionContainer: {
    padding: theme.spacing(1),
    flexShrink: 1,
    marginTop: theme.spacing(1),
    marginBottom: theme.spacing(1),
    display: 'inline-block',
  },
  breadCrumbsItem: {
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
  const classes = useStyles();
  const { t } = useTranslation();
  const { breadcrumbs, title } = props;

  return (
    <>
      <PageTitle title={title} />
      <Grid
        container
        alignItems="center"
        justifyContent="flex-start"
        className={classes.sectionContainer}
      >
        <Grid item>
          <Breadcrumbs
            aria-label={t('common|breadcrumbs').toString()}
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
          </Breadcrumbs>
        </Grid>
      </Grid>
    </>
  );
}
