import menuDown from '@iconify/icons-mdi/menu-down';
import menuUp from '@iconify/icons-mdi/menu-up';
import { InlineIcon } from '@iconify/react';
import Box from '@material-ui/core/Box';
import Button from '@material-ui/core/Button';
import Collapse from '@material-ui/core/Collapse';
import Divider from '@material-ui/core/Divider';
import Grid from '@material-ui/core/Grid';
import Link from '@material-ui/core/Link';
import Paper from '@material-ui/core/Paper';
import Table from '@material-ui/core/Table';
import TableBody from '@material-ui/core/TableBody';
import TableCell from '@material-ui/core/TableCell';
import TableHead from '@material-ui/core/TableHead';
import TableRow from '@material-ui/core/TableRow';
import Typography from '@material-ui/core/Typography';
import { makeStyles, useTheme } from '@material-ui/styles';
import { makeLocaleTime } from '../../constants/helpers';
import PropTypes from 'prop-types';
import React from "react";
import { Link as RouterLink } from 'react-router-dom';
import { instancesStore } from '../../stores/Stores';
import ChannelItem from '../Channels/Item.react';
import { CardFeatureLabel, CardLabel } from '../Common/Card';
import ListHeader from '../Common/ListHeader';
import Loader from '../Common/Loader';
import makeStatusDefs from './StatusDefs';

const useDetailsStyles = makeStyles(theme => ({
  timelineContainer: {
    maxHeight: '700px',
    overflow: 'auto',
  },
  divider: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  link: {
    fontSize: '1rem',
  }
}));

const useRowStyles = makeStyles(theme => ({
  statusExplanation: {
    padding: theme.spacing(2),
  }
}));

const useStatusStyles = makeStyles({
  statusButton: {
    textTransform: 'unset',
    verticalAlign: 'bottom'
  },
  // Align text with icon
  statusText: {
    display: 'inline',
    verticalAlign: 'bottom',
    lineHeight: '30px',
  },
});

function StatusLabel(props) {
  const classes = useStatusStyles();
  const statusDefs = makeStatusDefs(useTheme());
  let {status, activated} = props;
  let {icon=null, label='Unknown', color} = statusDefs[status.type] || {};
  const iconSize = '22px';

  return (
    <span>
      {icon !== null &&
        <InlineIcon
          icon={icon}
          color={color}
          width={iconSize}
          height={iconSize}
        />
      }
      {/* If there is no onClick passed to it, then we're not a button */}
      {props.onClick ?
        <Button
          size="small"
          onClick={props.onClick}
          className={classes.statusButton}
        >
          {label}
          <InlineIcon icon={ activated ? menuUp : menuDown } />
        </Button>
        :
        <Typography className={classes.statusText}>
          {label}
        </Typography>
      }
    </span>
  );
}

function StatusRow(props) {
  const classes = useRowStyles();
  let {entry} = props;
  let time = makeLocaleTime(entry.created_ts);
  const status = instancesStore.getInstanceStatus(entry.status, entry.version);
  let [collapsed, setCollapsed] = React.useState(true);

  function onStatusClick() {
    setCollapsed(!collapsed);
  }

  return (
    <React.Fragment>
      <TableRow>
        <TableCell>
          <StatusLabel
            onClick={onStatusClick}
            activated={!collapsed}
            status={status}
          />
        </TableCell>
        <TableCell>
          {entry.version}
        </TableCell>
        <TableCell>
          {time}
        </TableCell>
      </TableRow>
      <TableRow>
        <TableCell
          padding="none"
          colSpan={3}
        >
          <Collapse
            hidden={collapsed}
            in={!collapsed}
          >
            <Typography
              className={classes.statusExplanation}
            >
              {status.explanation}
            </Typography>
          </Collapse>
        </TableCell>
      </TableRow>
    </React.Fragment>
  );
}

function EventTable(props) {
  return (
    <Table>
      <TableHead>
        <TableCell>Status</TableCell>
        <TableCell>Version</TableCell>
        <TableCell>Time{props.events && props.events.status}</TableCell>
      </TableHead>
      <TableBody>
        {props.events.map((entry, i) =>
          <StatusRow key={i} entry={entry} />
        )
        }
      </TableBody>
    </Table>
  );
}

function DetailsView(props) {
  const classes = useDetailsStyles();
  let {application, group, instance} = props;
  let [eventHistory, setEventHistory] = React.useState([]);

  function onChangeInstances() {
    setEventHistory(instance.statusHistory || []);
  }

  React.useEffect(() => {
    onChangeInstances();
    instancesStore.addChangeListener(onChangeInstances);
    instancesStore.getInstanceStatusHistory(application.id, group.id, instance.id);

    return function cleanup() {
      instancesStore.removeChangeListener(onChangeInstances);
    };
  },
  [instance]);


  return (
    <Grid
      container
      spacing={1}
    >
      <Grid item md>
        <Paper>
          <ListHeader title="Instance Information" />
          <Box padding="1em">
            {application && group && instance &&
              <Grid container>
                <Grid item xs={12}>
                  <CardFeatureLabel>ID:</CardFeatureLabel>&nbsp;
                  <CardLabel>{instance.id}</CardLabel>
                </Grid>
                <Grid item xs={12}>
                  <CardFeatureLabel>IP:</CardFeatureLabel>&nbsp;
                  <CardLabel>{instance.ip}</CardLabel>
                </Grid>
                <Grid item xs={12}><Divider className={classes.divider} /></Grid>
                <Grid item xs={12}>
                  <CardFeatureLabel>Version:</CardFeatureLabel>&nbsp;
                  <CardLabel>{instance.application.version}</CardLabel>
                </Grid>
                <Grid item xs={12}>
                  <CardFeatureLabel>Status:</CardFeatureLabel>&nbsp;
                  <StatusLabel status={instance.statusInfo} />
                </Grid>
                <Grid item xs={12}>
                  <CardFeatureLabel>Last Update Check:</CardFeatureLabel>&nbsp;
                  {instance.statusInfo && <CardLabel>{makeLocaleTime(instance.statusInfo.created_ts)}</CardLabel>}
                </Grid>
                <Grid item xs={12}><Divider className={classes.divider} /></Grid>
                <Grid item xs={12}>
                  <CardFeatureLabel>Application:</CardFeatureLabel>&nbsp;
                  <Link className={classes.link} to={`/apps/${application.id}`} component={RouterLink}>{application.name}</Link>
                </Grid>
                <Grid item xs={12}>
                  <CardFeatureLabel>Group:</CardFeatureLabel>&nbsp;
                  <Link className={classes.link} to={`/apps/${application.id}/groups/${group.id}`} component={RouterLink}>{group.name}</Link>
                </Grid>
                <Grid item xs={12}>
                  <CardFeatureLabel>Channel:</CardFeatureLabel>
                  <ChannelItem channel={group.channel} ContainerComponent="span" />
                </Grid>
              </Grid>
            }
          </Box>
        </Paper>
      </Grid>
      <Grid item md>
        <Paper>
          <ListHeader title="Event Timeline" />
            {eventHistory ?
              <Box padding="1em">
                <div className={classes.timelineContainer}>
                  <EventTable events={eventHistory} />
                </div>
              </Box>
              :
              <Loader />
            }
        </Paper>
      </Grid>
    </Grid>
  );
}

DetailsView.propTypes = {
  application: PropTypes.object.isRequired,
  group: PropTypes.object.isRequired,
  instance: PropTypes.object,
};

export default DetailsView;
