import alertCircleOutline from '@iconify/icons-mdi/alert-circle-outline';
import alertOutline from '@iconify/icons-mdi/alert-outline';
import checkCircleOutline from '@iconify/icons-mdi/check-circle-outline';
import { Icon } from '@iconify/react';
import { Box } from '@material-ui/core';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import ListItem from '@material-ui/core/ListItem';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import { Activity } from '../../api/apiDataTypes';
import { activityStore } from '../../stores/Stores';

function ActivityItemIcon(props: { icon: object; color?: string | undefined }) {
  const { icon, color } = props;
  return <Icon icon={icon} color={color} width="30px" height="30px" />;
}

const stateIcons: {
  [key: string]: {
    icon: any;
    color: string;
  };
} = {
  warning: {
    icon: alertOutline,
    color: '#ff5500',
  },
  info: {
    icon: alertCircleOutline,
    color: '#00d3ff',
  },
  error: {
    icon: alertCircleOutline,
    color: '#F44336',
  },
  success: {
    icon: checkCircleOutline,
    color: '#22bb00',
  },
};

function Item(props: { entry: Activity }) {
  const [entryClass, setEntryClass] = React.useState<{ [key: string]: any }>({});
  const [entrySeverity, setEntrySeverity] = React.useState<{
    className?: string;
    [key: string]: any;
  }>({});

  function fetchEntryClassFromStore() {
    const entryClass = activityStore.getActivityEntryClass(props.entry.class, props.entry);
    setEntryClass(entryClass);
  }

  function fetchEntrySeverityFromStore() {
    const entrySeverity = activityStore.getActivityEntrySeverity(props.entry.severity);
    setEntrySeverity(entrySeverity);
  }

  React.useEffect(() => {
    fetchEntryClassFromStore();
    fetchEntrySeverityFromStore();
  }, []);

  const time = new Date(props.entry.created_ts).toLocaleString('default', {
    hour: '2-digit',
    minute: '2-digit',
  });
  let subtitle = '';
  let name: React.ReactNode = '';

  if (entryClass.type !== 'activityChannelPackageUpdated') {
    const { app_id, group_id } = props.entry;
    const groupPath = `apps/${app_id}/groups/${group_id}`;
    subtitle = 'GROUP';
    name = (
      <Link component={RouterLink} to={groupPath}>
        {entryClass.groupName}
      </Link>
    );
  }

  const stateIcon = stateIcons[entrySeverity.className || 'info'];

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
              <ActivityItemIcon {...stateIcon} />
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
                {entryClass.appName}
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
        {subtitle && (
          <Grid item container spacing={2} xs={12}>
            <Grid item>
              <Typography color="textSecondary" display="inline">
                {subtitle}
              </Typography>
            </Grid>
            <Grid item>{name}</Grid>
          </Grid>
        )}
        <Grid item>{entryClass.description}</Grid>
      </Grid>
    </ListItem>
  );
}

export default Item;
