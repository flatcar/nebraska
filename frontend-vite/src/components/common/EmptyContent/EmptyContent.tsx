import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
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
