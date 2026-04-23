import { Box } from '@mui/material';

import { useSelector } from '../stores/redux/hooks';

function Footer() {
  const { title = '', nebraska_version = '' } = useSelector(state => state.config);

  return (
    <Box
      sx={{
        mt: 1,
        color: 'text.secondary'
      }}>
      {`${title || 'Nebraska'} ${nebraska_version}`}
    </Box>
  );
}

export default Footer;
