import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

export default function Empty(props: { children: React.ReactNode }) {
  return (
    <Box
      sx={{
        padding: 2,
      }}
    >
      <Typography color="text.secondary" align="center" data-testid="empty">
        {props.children}
      </Typography>
    </Box>
  );
}
