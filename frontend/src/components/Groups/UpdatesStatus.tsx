import CheckIcon from '@mui/icons-material/Check';
import CloseIcon from '@mui/icons-material/Close';
import { Box } from '@mui/material';
import { useTranslation } from 'react-i18next';

export default function UpdatesStatus(props: { enabled: boolean; pausedHere?: boolean }) {
  const { t } = useTranslation();

  if (props.enabled) {
    return (
      <>
        <Box>{t('frequent|enabled')}</Box>
        <Box>
          <CheckIcon sx={{ color: theme => theme.palette.success.main }} fontSize="small" />
        </Box>
      </>
    );
  }

  return (
    <>
      <Box>{props.pausedHere ? t('groups|paused_here') : t('frequent|disabled')}</Box>
      <Box>
        <CloseIcon color="error" />
      </Box>
    </>
  );
}
