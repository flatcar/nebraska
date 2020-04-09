import alertOctagon from '@iconify/icons-mdi/alert-octagon';
import boxDownload from '@iconify/icons-mdi/box-download';
import checkboxMarked from '@iconify/icons-mdi/checkbox-marked';
import clipboardCheck from '@iconify/icons-mdi/clipboard-check';
import pauseCircle from '@iconify/icons-mdi/pause-circle';
import playCircle from '@iconify/icons-mdi/play-circle';
import progressDownload from '@iconify/icons-mdi/progress-download';
import questionMarkCircle from '@iconify/icons-mdi/question-mark-circle';

function makeStatusDefs(theme) {
  return {
    InstanceStatusComplete: {
      label: 'Complete',
      color: theme.palette.success.main,
      icon: clipboardCheck,
      queryValue: '4'
    },
    InstanceStatusDownloaded: {
      label: 'Downloaded',
      color: theme.palette.success['A700'],
      icon: boxDownload,
      queryValue: '6'
    },
    InstanceStatusOnHold: {
      label: 'On Hold',
      color: theme.palette.grey['400'],
      icon: pauseCircle,
      queryValue: '8'
    },
    InstanceStatusInstalled: {
      label: 'Installed',
      color: theme.palette.success['300'],
      icon: checkboxMarked,
      queryValue: '5'
    },
    InstanceStatusDownloading: {
      label: 'Downloading',
      color: theme.palette.success['700'],
      icon: progressDownload,
      queryValue: '7'
    },
    InstanceStatusError: {
      label: 'Error',
      color: theme.palette.error.main,
      icon: alertOctagon,
      queryValue: '3'
    },
    InstanceStatusUndefined: {
      label: 'Unknown',
      color: theme.palette.grey['500'],
      icon: questionMarkCircle,
      queryValue: '1'
    },
    InstanceStatusUpdateGranted: {
      label: 'Update Granted',
      color: theme.palette.success['500'],
      icon: playCircle,
      queryValue: '2'
    },
  };
}

export default makeStatusDefs;
