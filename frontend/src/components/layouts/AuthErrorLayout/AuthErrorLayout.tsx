import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import Typography from '@mui/material/Typography';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router';

import SectionPaper from '../../common/SectionPaper';

export default function AuthErrorLayout() {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();

  // Get error message from router state
  const errorMessage = location.state?.error || t('common|generic_error');

  return (
    <Grid container spacing={2} justifyContent="center" style={{ marginTop: '2rem' }}>
      <Grid
        size={{
          xs: 12,
          sm: 8,
          md: 6,
        }}
      >
        <SectionPaper>
          <Box textAlign="center" py={4}>
            <Typography variant="h5" gutterBottom>
              {t('common|authentication_error')}
            </Typography>
            <Typography variant="body1" color="textSecondary" paragraph>
              {errorMessage}
            </Typography>
            <Box mt={3}>
              <Button variant="outlined" onClick={() => navigate('/')}>
                {t('common|go_back')}
              </Button>
            </Box>
            <Typography variant="caption" color="textSecondary" display="block" mt={3}>
              {t('common|auth_error_help')}
            </Typography>
          </Box>
        </SectionPaper>
      </Grid>
    </Grid>
  );
}
