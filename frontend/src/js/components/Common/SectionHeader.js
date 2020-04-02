import Breadcrumbs from '@material-ui/core/Breadcrumbs';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import Paper from '@material-ui/core/Paper';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import ArrowLeftIos from '@material-ui/icons/ArrowBackIos';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

const useStyles = makeStyles(theme => ({
  sectionContainer: {
    padding: theme.spacing(1),
    flexShrink: 1,
    marginBottom: theme.spacing(1),
    display: 'inline-block',
  },
}));

export default function SectionHeader(props) {
  const classes = useStyles();
  const breadcrumbs = props.breadcrumbs;
  const title = props.title;

  return (
    <Paper elevation={0} className={classes.sectionContainer}>
      <Grid container alignItems="center" justify="flex-start">
        <Grid item>
          <Breadcrumbs aria-label="breadcrumbs">
            {breadcrumbs &&
              breadcrumbs.map(({path = null, label}, index) => {
                if (path)
                  return <Link to={path} component={RouterLink} key={index}>{label}</Link>;
                else
                  return <Typography key={index}>{label}</Typography>;
              }
              )}
            {title &&
              <Typography>{title}</Typography>
            }
          </Breadcrumbs>
        </Grid>
      </Grid>
    </Paper>
  );
}
