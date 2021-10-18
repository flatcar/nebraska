import { Box } from '@material-ui/core';
import React from 'react';
import { useSelector } from '../stores/redux/hooks';

function Footer() {
  const { title = '', nebraska_version = '' } = useSelector(state => state.config);

  return (
    <Box mt={1} color="text.secondary">
      {`${title || 'Nebraska'} ${nebraska_version}`}
    </Box>
  );
}

export default Footer;
