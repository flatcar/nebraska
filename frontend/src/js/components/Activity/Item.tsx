import alertCircleOutline from '@iconify/icons-mdi/alert-circle-outline';
import alertOutline from '@iconify/icons-mdi/alert-outline';
import checkCircleOutline from '@iconify/icons-mdi/check-circle-outline';
import { Icon } from '@iconify/react';
import { Box, makeStyles } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import ListItem from '@material-ui/core/ListItem';
import Typography from '@material-ui/core/Typography';
import PropTypes from 'prop-types';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
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
      const {app_id, group_id} = this.props.entry;
      const groupPath = `apps/${app_id}/groups/${group_id}`;
      subtitle = 'GROUP';
      name = <Link component={RouterLink} to={groupPath}>{this.state.entryClass.groupName}</Link>;
    }

    const stateIcon = stateIcons[this.state.entrySeverity.className || 'info'];

    return (
      <ListItem
        alignItems="flex-start"
        disableGutters
        style={{
          paddingTop: '15px',
          paddingLeft: '15px',
        }}
      >
        <Grid container alignItems="center" justify="space-between">
          <Grid item xs={10}>
            <Box display="flex" alignItems="center" justifyContent="flex-start">
              <Box mr={1}>
                <ActivityItemIcon {...stateIcon} time={time} />
              </Box>
              <Box>
                <Typography
                  // @todo: Move this into a classes object once we convert this component to a
                  // functional one.
                  style={{
                    fontWeight: 'bold',
                    fontSize: '1.1rem',
                    color: '#474747',
                  }}
                >
                  {this.state.entryClass.appName}
                </Typography>
              </Box>
            </Box>
          </Grid>

          <Grid item xs={2}>
            <Typography
              color="textSecondary"
              // @todo: Move this into a classes object once we convert this component to a
              // functional one.
              style={{
                fontSize: '.7rem',
              }}
            >
              {time}
            </Typography>
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
