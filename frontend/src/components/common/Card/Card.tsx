import { Box } from '@mui/material';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Link from '@mui/material/Link';
import { useTheme } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import makeStyles from '@mui/styles/makeStyles';
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
    fontWeight: 'bold',
  },
  mainLink: {
    fontSize: '1.8rem',
    fontWeight: 'bold',
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
    fontSize: '1rem',
  },
  innerDivider: {
    marginLeft: '1em',
    marginRight: '1em',
  },
  label: props => ({
    fontSize: '1rem',
    ...props,
  }),
}));

export interface CardFeatureLabelProps {
  children: React.ReactNode;
}

export function CardFeatureLabel(props: CardFeatureLabelProps) {
  const classes = useStyles();
  return (
    <Typography component="span" className={classes.featureLabel}>
      {props.children}
    </Typography>
  );
}

export interface CardDescriptionLabelProps {
  children: React.ReactNode;
}

export function CardDescriptionLabel(props: CardDescriptionLabelProps) {
  const classes = useStyles();
  return (
    <Box mt={2}>
      <Typography component="span" className={classes.descriptionLabel}>
        {props.children}
      </Typography>
    </Box>
  );
}

export interface CardLabelProps {
  children: React.ReactNode;
  labelStyle?: object;
}

export function CardLabel(props: CardLabelProps) {
  const { labelStyle = {} } = props;
  const classes = useStyles(labelStyle);
  return (
    <Typography component="span" className={classes.label}>
      {props.children}
    </Typography>
  );
}

export interface CardHeaderProps {
  cardMainLinkPath?: string | { pathname: string };
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
      <Grid container className={classes.gridHeader} justifyContent="space-between">
        <Grid item container spacing={1} alignItems="center" justifyContent="space-between">
          <Grid item xs={12}>
            {props.cardMainLinkPath ? (
              <Typography variant="h2">
                <Link
                  component={RouterLink}
                  to={props.cardMainLinkPath}
                  className={classes.mainLink}
                >
                  {props.cardMainLinkLabel}
                </Link>
              </Typography>
            ) : (
              <Typography className={classes.cardTitle}>{props.cardMainLinkLabel}</Typography>
            )}
          </Grid>
          <Grid item xs={12}>
            <Box bgcolor={theme.palette.lightSilverShade} px={1} display="inline-block">
              <Typography className={classes.idLabel} arial-label="group-id" noWrap>
                {props.cardTrack ? props.cardTrack : props.cardId}
              </Typography>
            </Box>
            <CardDescriptionLabel>{props.cardDescription}</CardDescriptionLabel>
          </Grid>
        </Grid>
        {props.children && <Grid item>{props.children}</Grid>}
      </Grid>
      <Divider light className={classes.innerDivider} />
    </React.Fragment>
  );
}
