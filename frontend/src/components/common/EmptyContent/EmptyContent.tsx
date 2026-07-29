import Box from '@mui/material/Box';
import Typography from '@mui/material/Typography';
import React from 'react';

export default function Empty(props: { children: React.ReactNode }) {
  return (
    <Box
      padding={3}
      sx={{
        borderRadius: 2,
        border: '1px dashed #D0D5DD',
        backgroundColor: 'rgba(244, 246, 248, 0.8)',
      }}
    >
      <Typography
        color="textSecondary"
        align="center"
        data-testid="empty"
        sx={{ lineHeight: 1.6, fontSize: '0.9rem' }}
      >
        {props.children}
      </Typography>
    </Box>
  );
}
