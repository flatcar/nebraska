import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import React from 'react';
import Typography from '@material-ui/core/Typography';
import { Link as RouterLink } from "react-router-dom"
import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  gridHeader: {
    background: '#fafafa',
    padding: '1em',
  },
  mainLink: {
    color: 'rgba(119,119,119)',
    fontSize: 25,
    fontWeight: 'bold',
  },
  featureLabel: {
    color: 'rgb(119,119,119, .9)',
    textTransform: 'uppercase',
    fontSize: 14,
  },
  descriptionLabel: {
    color: 'rgb(119,119,119, .9)',
    fontSize: 16,
  },
  idLabel: {
    color: 'rgba(119,119,119,.75)',
    fontSize: 16
  },
}));

export function CardFeatureLabel(props) {
  const classes = useStyles();
  return (
    <Typography component='span' className={classes.featureLabel}>{props.children}</Typography>
  );
}

export function CardDescriptionLabel(props) {
  const classes = useStyles();
  return (
    <Typography component='span' className={classes.descriptionLabel}>{props.children}</Typography>
  );
}

export function CardHeader(props) {
  const classes = useStyles();
  return (
    <Grid
      container
      className={classes.gridHeader}
      justify="space-between"
    >
      <Grid item xs={11}>
        <Grid
          container
          spacing={1}
        >
          <Grid item xs={6}>
            { props.cardMainLinkPath ? (
              <Link component={RouterLink} button to={props.cardMainLinkPath} className={classes.mainLink}>
                {props.cardMainLinkLabel}
              </Link>
            ) : (
              <Typography className={classes.mainLink}>
                {props.cardMainLinkLabel}
              </Typography>
            )
            }
          </Grid>
          <Grid item xs={6}>
            <Typography className={classes.idLabel} arial-label="group-id">{props.cardId}</Typography>
          </Grid>
          <Grid item xs={12}>
            <CardDescriptionLabel variant="h5">{props.cardDescription}</CardDescriptionLabel>
          </Grid>
        </Grid>
      </Grid>
      {props.children &&
        <Grid item>
          {props.children}
        </Grid>
      }
    </Grid>
  );
}
