import alertCircleOutline from '@iconify/icons-mdi/alert-circle-outline';
import alertOutline from '@iconify/icons-mdi/alert-outline';
import checkCircleOutline from '@iconify/icons-mdi/check-circle-outline';
import { Icon } from '@iconify/react';
import { Box, makeStyles } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import Typography from '@material-ui/core/Typography';
import PropTypes from 'prop-types';
import React from 'react';
import { activityStore } from '../../stores/Stores';

const useStyles = makeStyles(theme => ({
  timeText: {
    color: theme.palette.text.secondary,
    fontSize: '.9em',
  },
}));

function ActivityItemIcon(props) {
  const {icon, color} = props;
  return (
    <Icon
      icon={icon}
      color={color}
      width="30px"
      height="30px"
    />
  );
}

const stateIcons = {
  warning: {
    icon: alertOutline,
    color: '#ff5500'
  },
  info: {
    icon: alertCircleOutline,
    color: '#00d3ff'
  },
  error: {
    icon: alertCircleOutline,
    color: '#F44336'
  },
  success: {
    icon: checkCircleOutline,
    color: '#22bb00'
  },
};

class Item extends React.Component {

  constructor(props) {
    super(props);

    this.state = {
      entryClass: {},
      entrySeverity: {}
    };
  }

  fetchEntryClassFromStore() {
    const entryClass = activityStore
      .getActivityEntryClass(this.props.entry.class, this.props.entry);
    this.setState({
      entryClass: entryClass
    });
  }

  fetchEntrySeverityFromStore() {
    const entrySeverity = activityStore.getActivityEntrySeverity(this.props.entry.severity);
    this.setState({
      entrySeverity: entrySeverity
    });
  }

  componentDidMount() {
    this.fetchEntryClassFromStore();
    this.fetchEntrySeverityFromStore();
  }

  render() {
    const time = new Date(this.props.entry.created_ts).toLocaleString('default', {hour: '2-digit', minute: '2-digit'});
    let subtitle = '';
    let name = '';

    if (this.state.entryClass.type !== 'activityChannelPackageUpdated') {
      subtitle = 'GROUP';
      name = this.state.entryClass.groupName;
    }

    const stateIcon = stateIcons[this.state.entrySeverity.className || 'info'];

    return (
      <ListItem alignItems="flex-start">
        <Grid container alignItems="center" justify="space-between">
          <Grid item xs={10}>
            <Box display="flex" alignItems="center" justifyContent="flex-start">
              <Box mr={1}>
                <ActivityItemIcon {...stateIcon} time={time} />
              </Box>
              <Box>
                <Typography variant="h6">{this.state.entryClass.appName}</Typography>
              </Box>
            </Box>
          </Grid>

          <Grid item xs={2}>
            <Typography color="textSecondary" variant="subtitle1">{time}</Typography>
          </Grid>
          {subtitle && <Grid item container spacing={2} xs={12}>
            <Grid item>
              <Typography color="textSecondary" display="inline">
                {subtitle}
              </Typography>
            </Grid>
            <Grid item>
              {name}
            </Grid>
          </Grid>
          }
          <Grid item>
            {this.state.entryClass.description}
          </Grid>
        </Grid>
      </ListItem>
    );
  }

}

Item.propTypes = {
  entry: PropTypes.object.isRequired
};

export default Item;
