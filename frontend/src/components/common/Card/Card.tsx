import { Box } from '@mui/material';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Link from '@mui/material/Link';
import { useTheme } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';

export interface CardFeatureLabelProps {
  children: React.ReactNode;
}

export function CardFeatureLabel(props: CardFeatureLabelProps) {
  return (
    <Typography
      component="span"
      sx={{
        color: theme => theme.palette.text.secondary,
        textTransform: 'uppercase',
        fontSize: '1rem',
      }}
    >
      {props.children}
    </Typography>
  );
}

export interface CardDescriptionLabelProps {
  children: React.ReactNode;
}

export function CardDescriptionLabel(props: CardDescriptionLabelProps) {
  return (
    <Box mt={2}>
      <Typography
        component="span"
        sx={{
          color: theme => theme.palette.text.secondary,
          fontSize: '0.875rem',
        }}
      >
        {props.children}
      </Typography>
    </Box>
  );
}

export interface CardLabelProps {
  children: React.ReactNode;
  labelStyle?: object;
}

export function CardLabel({ children, labelStyle }: CardLabelProps) {
  return (
    <Typography
      component="span"
      sx={{
        fontSize: '1rem',
        ...labelStyle,
      }}
    >
      {children}
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
  const theme = useTheme();
  return (
    <React.Fragment>
      <Grid
        container
        sx={{
          padding: '1rem',
          flexWrap: 'nowrap',
        }}
        justifyContent="space-between"
      >
        <Grid container spacing={1} alignItems="center" justifyContent="space-between">
          <Grid size={12}>
            {props.cardMainLinkPath ? (
              <Typography variant="h2">
                <Link
                  component={RouterLink}
                  to={props.cardMainLinkPath}
                  sx={{
                    fontSize: '1.8rem',
                    fontWeight: 'bold',
                  }}
                  underline="hover"
                >
                  {props.cardMainLinkLabel}
                </Link>
              </Typography>
            ) : (
              <Typography
                sx={{
                  color: '#474747',
                  fontSize: '1.8rem',
                  fontWeight: 'bold',
                }}
              >
                {props.cardMainLinkLabel}
              </Typography>
            )}
          </Grid>
          <Grid size={12}>
            <Box bgcolor={theme.palette.lightSilverShade} px={1} display="inline-block">
              <Typography
                sx={{
                  color: theme => theme.palette.text.secondary,
                  fontSize: '1rem',
                }}
                arial-label="group-id"
                noWrap
              >
                {props.cardTrack ? props.cardTrack : props.cardId}
              </Typography>
            </Box>
            <CardDescriptionLabel>{props.cardDescription}</CardDescriptionLabel>
          </Grid>
        </Grid>
        {props.children && <Grid>{props.children}</Grid>}
      </Grid>
      <Divider
        light
        sx={{
          marginLeft: '1em',
          marginRight: '1em',
        }}
      />
    </React.Fragment>
  );
}
