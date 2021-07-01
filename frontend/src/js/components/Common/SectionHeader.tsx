import { Box } from '@material-ui/core';
import Breadcrumbs from '@material-ui/core/Breadcrumbs';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import NavigateNextIcon from '@material-ui/icons/NavigateNext';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router-dom';

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
  const breadcrumbs = props.breadcrumbs;
  const title = props.title;
  const { t } = useTranslation();

  return (
    <Grid container alignItems="center" justify="flex-start" className={classes.sectionContainer}>
      <Grid item>
        <Breadcrumbs
          aria-label={t('common|breadcrumbs')}
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
                    <Link to={path} component={RouterLink}>
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
  );
}
