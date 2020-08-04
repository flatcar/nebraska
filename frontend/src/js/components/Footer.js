import { Box } from '@material-ui/core';
import React from 'react';
import API from '../api/API';

function Footer() {
  const [nebraskaVersion, setNebraskaVersion] = React.useState(null);
  React.useEffect(() => {
    API.getConfig().then((config) => {
      setNebraskaVersion(config.nebraska_version);
    });
  });
  return <Box mt={1} color="text.secondary">{nebraskaVersion && `Nebraska v${nebraskaVersion}`}</Box>;
}

export default Footer;
