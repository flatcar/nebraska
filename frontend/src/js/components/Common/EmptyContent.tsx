import Box from '@material-ui/core/Box';
import Typography from '@material-ui/core/Typography';
import React from 'react';

export default function Empty(props: { children: React.ReactNode }) {
  return (
    <Box padding={2}>
      <Typography color="textSecondary" align="center" data-testid="empty">
        {props.children}
      </Typography>
    </Box>
  );
}
