import { Box } from '@mui/material';

import { useSelector } from '../stores/redux/hooks';

function Footer() {
  const { title = '', nebraska_version = '' } = useSelector(state => state.config);

  return (
    <Box
      mt={3}
      mb={2}
      color="text.secondary"
      fontSize="0.75rem"
      textAlign="center"
      sx={{ opacity: 0.85 }}
    >
      {`${title || 'Nebraska'} ${nebraska_version}`}
    </Box>
  );
}

export default Footer;
