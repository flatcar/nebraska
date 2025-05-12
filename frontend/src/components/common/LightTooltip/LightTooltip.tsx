import { Tooltip } from '@mui/material';
import withStyles from '@mui/styles/withStyles';

const LightTooltip = withStyles(theme => ({
  tooltip: {
    backgroundColor: theme.palette.common.white,
    color: 'rgba(0, 0, 0, 0.87)',
    boxShadow: theme.shadows[1],
    fontSize: '1rem',
    whiteSpace: 'pre-line',
  },
}))(Tooltip);
export default LightTooltip;
