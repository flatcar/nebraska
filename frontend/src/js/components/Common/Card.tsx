import { Box } from '@material-ui/core';
import Divider from '@material-ui/core/Divider';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import { makeStyles, useTheme } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

const useStyles = makeStyles(theme => ({
  gridHeader: {
    padding: '1rem',
    flexWrap: 'nowrap',
  },
  cardTitle: {
    color: '#474747',
    fontSize: '1.8rem',
    fontWeight: 'bold'
  },
  mainLink: {
    fontSize: '1.8rem',
    fontWeight: 'bold'
  },
  featureLabel: {
    color: theme.palette.text.secondary,
    textTransform: 'uppercase',
    fontSize: '1rem',
  },
  descriptionLabel: {
    color: theme.palette.text.secondary,
    fontSize: '0.875rem',
  },
  idLabel: {
    color: theme.palette.text.secondary,
    fontSize: '1rem'
  },
  innerDivider: {
    marginLeft: '1em',
    marginRight: '1em',
  },
  label: props => ({
    fontSize: '1rem',
    ...props
  }),
}));

export function CardFeatureLabel(props: {children: React.ReactNode}) {
  const classes = useStyles();
  return (
    <Typography component='span' className={classes.featureLabel}>{props.children}</Typography>
  );
}

export function CardDescriptionLabel(props: {children: React.ReactNode}) {
  const classes = useStyles();
  return (
    <Box mt={2}>
      <Typography component='span' className={classes.descriptionLabel}>{props.children}</Typography>
    </Box>
  );
}

export function CardLabel(props: {children: React.ReactNode; labelStyle?: object}) {
  const {labelStyle = {}} = props;
  const classes = useStyles(labelStyle);
  return (
    <Typography component='span' className={classes.label}>{props.children}</Typography>
  );
}

interface CardHeaderProps {
  cardMainLinkPath?: string | {pathname: string};
  cardMainLinkLabel?: string;
  cardTrack?: string;
  cardId?: string;
  cardDescription: React.ReactNode;
  children?: React.ReactNode;
}

export function CardHeader(props: CardHeaderProps) {
  const classes = useStyles();
  const theme = useTheme();
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
          justify="space-between"
        >
          <Grid item xs={12}>
            { props.cardMainLinkPath ? (
              <Link component={RouterLink} to={props.cardMainLinkPath} className={classes.mainLink}>
                {props.cardMainLinkLabel}
              </Link>
            ) : (
              <Typography className={classes.cardTitle}>
                {props.cardMainLinkLabel}
              </Typography>
            )
            }
          </Grid>
          <Grid item xs={12}>
            <Box bgcolor={theme.palette.lightSilverShade} px={1} display="inline-block">
              <Typography className={classes.idLabel} arial-label="group-id" noWrap>{props.cardTrack ? props.cardTrack : props.cardId}</Typography>
            </Box>
            <CardDescriptionLabel>{props.cardDescription}</CardDescriptionLabel>
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
