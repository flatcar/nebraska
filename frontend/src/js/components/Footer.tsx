import { Box } from '@material-ui/core';
import React from 'react';
import { useTypedSelector } from '../stores/redux/reducers';

function Footer() {
  const {title='', nebraska_version=''} = useTypedSelector(state => state.config);

  return (
    <Box mt={1} color="text.secondary">
      {`${title || 'Nebraska'} ${nebraska_version}`}
    </Box>
  );
}

export default Footer;
