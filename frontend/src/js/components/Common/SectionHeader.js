import { Box } from '@material-ui/core';
import Breadcrumbs from '@material-ui/core/Breadcrumbs';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import Paper from '@material-ui/core/Paper';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import NavigateNextIcon from '@material-ui/icons/NavigateNext';
import React from 'react';
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
      color: theme.palette.text.secondary
    }
  }
}));

export default function SectionHeader(props) {
  const classes = useStyles();
  const breadcrumbs = props.breadcrumbs;
  const title = props.title;

  return (
    <Grid container alignItems="center" justify="flex-start" className={classes.sectionContainer}>
      <Grid item>
        <Breadcrumbs aria-label="breadcrumbs" separator={<NavigateNextIcon fontSize="small" />}>
          {breadcrumbs &&
              breadcrumbs.map(({path = null, label}, index) => {
                if (path)
                  return (
                    <Box component="span" className={classes.breadCrumbsItem}>
                      <Link to={path} component={RouterLink} key={index}>
                        {label}
                      </Link>
                    </Box>
                  );
                else
                  return <Typography key={index} color="textPrimary">{label}</Typography>;
              }
              )}
          {title &&
          <Typography color="textPrimary">{title}</Typography>
          }
        </Breadcrumbs>
      </Grid>
    </Grid>
  );
}
