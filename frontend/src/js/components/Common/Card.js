import Divider from '@material-ui/core/Divider';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import { makeStyles } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import { Link as RouterLink } from "react-router-dom";

const useStyles = makeStyles(theme => ({
  gridHeader: {
    padding: '1rem',
    flexWrap: 'nowrap',
  },
  mainLink: {
    fontSize: '2.2rem',
  },
  featureLabel: {
    color: theme.palette.text.secondary,
    textTransform: 'uppercase',
    fontSize: '1rem',
  },
  descriptionLabel: {
    color: theme.palette.text.secondary,
    fontSize: '1rem',
  },
  idLabel: {
    color: theme.palette.text.secondary,
    fontSize: '1rem'
  },
  innerDivider: {
    marginLeft: '1em',
    marginRight: '1em',
  }
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
    <React.Fragment>
      <Grid
        container
        className={classes.gridHeader}
        justify="space-between"
      >
        <Grid item
          container
          spacing={1}
          alignItems="center"
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
        {props.children &&
          <Grid item>
            {props.children}
          </Grid>
        }
      </Grid>
      <Divider light className={classes.innerDivider} />
    </React.Fragment>
  );
}
