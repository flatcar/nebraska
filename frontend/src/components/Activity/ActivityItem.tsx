import { Box } from '@mui/material';
import Grid from '@mui/material/Grid';
import Link from '@mui/material/Link';
import ListItem from '@mui/material/ListItem';
import Typography from '@mui/material/Typography';
import makeStyles from '@mui/styles/makeStyles';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router-dom';

import { Activity } from '../../api/apiDataTypes';
import { toLocaleString } from '../../i18n/dateTime';
import { activityStore } from '../../stores/Stores';
import ActivityItemIcon from './ActivityItemIcon';

const useStyles = makeStyles({
  groupLink: {
    color: '#1b5c91',
  },
  appName: {
    fontWeight: 'bold',
    fontSize: '1.1rem',
    color: '#474747',
  },
  time: {
    fontSize: '.7rem',
  },
  list: {
    paddingTop: '15px',
    paddingLeft: '15px',
  },
});

export interface ActivityItemProps {
  entry: Activity;
}

export default function ActivityItem(props: ActivityItemProps) {
  const entryClass = activityStore().makeActivityEntryClass(props.entry.class, props.entry);
  const entrySeverity = activityStore().makeActivityEntrySeverity(props.entry.severity);

  return (
    <ActivityItemPure
      createdTs={props.entry.created_ts}
      appId={props.entry.app_id}
      groupId={props.entry.group_id}
      classType={entryClass.type}
      groupName={entryClass.groupName}
      appName={entryClass.appName}
      description={entryClass.description}
      severityName={entrySeverity.className}
    />
  );
}

export interface ActivityItemPureProps {
  appId: string;
  appName: string;
  classType: string;
  createdTs: string;
  description: string | React.ReactElement<any, string | React.JSXElementConstructor<any>>;
  groupId: string;
  groupName: string | null;
  severityName: string;
}

export function ActivityItemPure(props: ActivityItemPureProps) {
  const classes = useStyles();
  const { t } = useTranslation();

  const time = toLocaleString(props.createdTs, undefined, {
    hour: '2-digit',
    minute: '2-digit',
  });
  let subtitle = '';
  let name: React.ReactNode = '';

  if (props.classType !== 'activityChannelPackageUpdated') {
    const groupPath = `apps/${props.appId}/groups/${props.groupId}`;
    subtitle = t('activity|group');
    name = (
      <Link component={RouterLink} to={groupPath} className={classes.groupLink} underline="hover">
        {props.groupName}
      </Link>
    );
  }

  return (
    <ListItem alignItems="flex-start" disableGutters className={classes.list}>
      <Grid container alignItems="center" justifyContent="space-between">
        <Grid size={10}>
          <Box display="flex" alignItems="center" justifyContent="flex-start">
            <Box mr={1}>
              <ActivityItemIcon severityName={props.severityName} />
            </Box>
            <Box>
              <Typography className={classes.appName}>{props.appName}</Typography>
            </Box>
          </Box>
        </Grid>
        <Grid size={2}>
          <Typography color="textSecondary" className={classes.time}>
            {time}
          </Typography>
        </Grid>
        {subtitle && (
          <Grid container spacing={2} size={12}>
            <Grid>
              <Typography color="textSecondary" display="inline">
                {subtitle}
              </Typography>
            </Grid>
            <Grid>{name}</Grid>
          </Grid>
        )}
        <Grid>{props.description}</Grid>
      </Grid>
    </ListItem>
  );
}
