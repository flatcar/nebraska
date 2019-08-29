import Breadcrumbs from '@material-ui/core/Breadcrumbs';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import ArrowLeftIos from '@material-ui/icons/ArrowBackIos';
import React from 'react';
import { Link as RouterLink } from "react-router-dom";

const useStyles = makeStyles(theme => ({
  sectionContainer: {
    paddingTop: '1em',
    paddingBottom: '1.5em',
  }
}));

export default function SectionHeader(props) {
  const classes = useStyles();
  let breadcrumbs = props.breadcrumbs;
  let title = props.title;

  return (
    <Grid
      container
      alignItems="center"
      className={classes.sectionContainer}
    >
      <Grid item xs={3}>
        {breadcrumbs &&
          <Grid container alignItems="center">
            <Grid item>
              <ArrowLeftIos fontSize="inherit"/>
            </Grid>
            <Grid item>
          <Breadcrumbs aria-label="breadcrumbs">
            {breadcrumbs.map(({path=null, label}) => {
              if (path)
                return <Link to={path} component={RouterLink}>{label}</Link>;
              else
                return <Typography>{label}</Typography>;
              }
            )}
          </Breadcrumbs>
          </Grid>
        </Grid>
        }
      </Grid>
      <Grid item xs={6}>
        <Typography variant="h3" align="center" color="primary">
          {title}
        </Typography>
      </Grid>
      <Grid item xs={3} aria-hidden="true"></Grid>
    </Grid>
  );
}