import { Box } from '@mui/material';
import Grid from '@mui/material/Grid';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import React from 'react';

const PREFIX = 'ListHeader';

const classes = {
  sectionHeader: `${PREFIX}-sectionHeader`
};

const StyledGrid = styled(Grid)({
  [`&.${classes.sectionHeader}`]: {
    padding: '1em',
  },
});

export default function ListHeader(props: { title: string; actions?: React.ReactElement[] }) {

  const actions = props.actions || [];
  return (
    <StyledGrid
      container
      alignItems="flex-start"
      justifyContent="space-between"
      className={classes.sectionHeader}
    >
      {props.title && (
        <Grid>
          <Box p={0.5}>
            <Typography variant="h1">{props.title}</Typography>
          </Box>
        </Grid>
      )}
      {actions && actions.map((action, i: number) => <Grid key={i}>{action}</Grid>)}
    </StyledGrid>
  );
}
