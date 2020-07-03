import chevronDown from '@iconify/icons-mdi/chevron-down';
import chevronUp from '@iconify/icons-mdi/chevron-up';
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
import PropTypes from 'prop-types';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import API from '../../api/API';
import { getInstanceStatus, makeLocaleTime } from '../../constants/helpers';
import ChannelItem from '../Channels/Item.react';
import { CardFeatureLabel, CardLabel } from '../Common/Card';
import Empty from '../Common/EmptyContent';
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
  },
  root: {
    '& .MuiTableCell-root': {
      padding: '0.5rem'
    }
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
  const {status, activated} = props;
  const {icon = null, label = 'Unknown', color} = status && statusDefs[status.type] || {};
  const iconSize = '22px';

  return (
    <span>
      {/* If there is no onClick passed to it, then we're not a button */}
      {props.onClick ?
        <Button
          size="small"
          onClick={props.onClick}
          className={classes.statusButton}
        >
          <Box bgcolor={status.bgColor} color={status.textColor} p={0.8} display="inline-block" mr={1}>
            {label}
          </Box>
          <InlineIcon icon={ activated ? chevronUp : chevronDown }
            height="25"
            width="25"
            color="#808080"
          />
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
  const {entry} = props;
  const time = makeLocaleTime(entry.created_ts);
  const status = getInstanceStatus(entry.status, entry.version);
  const [collapsed, setCollapsed] = React.useState(true);

  function onStatusClick() {
    setCollapsed(!collapsed);
  }

  return (
    <React.Fragment>
      <TableRow className={classes.root}>
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
  return props.events.length === 0 ? (
    <Empty>
      No events to report for this instance yet.
    </Empty>)
    : (
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Status</TableCell>
            <TableCell>Version</TableCell>
            <TableCell>Time{props.events && props.events.status}</TableCell>
          </TableRow>
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
  const theme = useTheme();
  const {application, group, instance} = props;
  const [eventHistory, setEventHistory] = React.useState([]);
  React.useEffect(() => {
    API.getInstanceStatusHistory(application.id, group.id, instance.id).then((statusHistory) => {
      setEventHistory(statusHistory);
    })
      .catch(() => {
        setEventHistory([]);
      });
  },
  [instance]);

  return (
    <>
      <ListHeader title="Instance Information" />
      <Paper>
        <Box p={2}>
          <Grid
            container
            spacing={1}
          >
            <Grid item xs={12}>
              <Box fontWeight={700} fontSize={30} color={theme.palette.greyShadeColor}>
                {instance.id}
              </Box>
            </Grid>
            <Grid item md>

              <Box mt={2}>
                {application && group && instance &&
                <Grid container>
                  <Grid item xs={12} container>
                    <Grid item container>
                      <Grid item xs={6}>
                        <CardFeatureLabel>IP</CardFeatureLabel>
                        <Box mt={1}>
                          <CardLabel>{instance.ip}</CardLabel>
                        </Box>
                      </Grid>
                      <Grid item xs={6}>
                        <CardFeatureLabel>Version</CardFeatureLabel>
                        <Box mt={1}>
                          <CardLabel>{instance.application.version}</CardLabel>
                        </Box>
                      </Grid>
                    </Grid>
                    <Grid item xs={12}><Divider className={classes.divider} /></Grid>

                    <Grid item xs={12} container>
                      <Grid item xs={6}>
                        <CardFeatureLabel>Status</CardFeatureLabel>
                        <Box mt={1}>
                          <StatusLabel
                            status={instance.statusInfo}
                          />
                        </Box>
                      </Grid>
                      <Grid item xs={6}>
                        <CardFeatureLabel>Last Update Check</CardFeatureLabel>
                        <Box mt={1}>
                          <CardLabel>
                            {makeLocaleTime(instance.application.last_check_for_updates)}
                          </CardLabel>
                        </Box>
                      </Grid>
                    </Grid>

                    <Grid item xs={12}><Divider className={classes.divider} /></Grid>
                    <Grid item xs={12} container>
                      <Grid item xs={6}>
                        <CardFeatureLabel>Application</CardFeatureLabel>
                        <Box mt={1}>
                          <Link className={classes.link} to={`/apps/${application.id}`} component={RouterLink}>{application.name}</Link>
                        </Box>
                      </Grid>
                      <Grid item xs={6}>
                        <CardFeatureLabel>Group</CardFeatureLabel>
                        <Box mt={1}>
                          <Link className={classes.link} to={`/apps/${application.id}/groups/${group.id}`} component={RouterLink}>{group.name}</Link>
                        </Box>
                      </Grid>
                    </Grid>

                    <Grid item xs={12}>
                      <Box mt={2}>
                        <CardFeatureLabel>Channel</CardFeatureLabel>&nbsp;
                        {group.channel ?
                          <ChannelItem channel={group.channel} />
                          :
                          <CardLabel>None</CardLabel>
                        }
                      </Box>
                    </Grid>
                  </Grid>

                </Grid>
                }
              </Box>
            </Grid>
            <Box width="1%">
              <Divider orientation="vertical" variant="fullWidth"/>
            </Box>
            <Grid item md>
              <Box mt={2} fontSize={18} fontWeight={700} color={theme.palette.greyShadeColor}>
                Event Timeline
                {eventHistory ?
                  <Box padding="1em">
                    <div className={classes.timelineContainer}>
                      <EventTable events={eventHistory} />
                    </div>
                  </Box>
                  :
                  <Loader />
                }
              </Box>
            </Grid>
          </Grid>
        </Box>
      </Paper>
    </>
  );
}

DetailsView.propTypes = {
  application: PropTypes.object.isRequired,
  group: PropTypes.object.isRequired,
  instance: PropTypes.object,
};

export default DetailsView;
