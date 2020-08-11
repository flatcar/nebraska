import { Box } from '@material-ui/core';
import React from 'react';
import API from '../api/API';

function Footer() {
  const [nebraskaConfig, setNebraskaConfig] = React.useState(null);
  React.useEffect(() => {
    API.getConfig().then((config) => {
      setNebraskaConfig(config);
    });
  }, []);
  return <Box mt={1} color="text.secondary">{nebraskaConfig && `${nebraskaConfig.title || 'Nebraska'} ${nebraskaConfig.nebraska_version}`}</Box>;
}

export default Footer;
