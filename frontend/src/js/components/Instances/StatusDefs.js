import alertOctagon from '@iconify/icons-mdi/alert-octagon';
import progressDownload from '@iconify/icons-mdi/progress-download';
import boxDownload from '@iconify/icons-mdi/box-download';
import clipboardCheck from '@iconify/icons-mdi/clipboard-check';
import checkboxMarked from '@iconify/icons-mdi/checkbox-marked';
import pauseCircle from '@iconify/icons-mdi/pause-circle';
import questionMarkCircle from '@iconify/icons-mdi/question-mark-circle';
import playCircle from '@iconify/icons-mdi/play-circle';

function makeStatusDefs(theme) {
  return {
    InstanceStatusComplete: {
      label: 'Complete',
      color: theme.palette.success.main,
      icon: clipboardCheck,
    },
    InstanceStatusDownloaded: {
      label: 'Downloaded',
      color: theme.palette.success['A700'],
      icon: boxDownload,
    },
    InstanceStatusOnHold: {
      label: 'On Hold',
      color: theme.palette.grey['500'],
      icon: pauseCircle,
    },
    InstanceStatusInstalled: {
      label: 'Installed',
      color: theme.palette.success['400'],
      icon: checkboxMarked,
    },
    InstanceStatusDownloading: {
      label: 'Downloading',
      color: theme.palette.success['A700'],
      icon: progressDownload,
    },
    InstanceStatusError: {
      label: 'Error',
      color: theme.palette.error.main,
      icon: alertOctagon,
    },
    InstanceStatusUndefined: {
      label: 'Unknown',
      color: theme.palette.grey['500'],
      icon: questionMarkCircle,
    },
    InstanceStatusUpdateGranted: {
      label: 'Granted',
      color: theme.palette.success['400'],
      icon: playCircle,
    },
  };
}

export default makeStatusDefs;
