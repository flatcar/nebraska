import chevronDown from '@iconify/icons-mdi/chevron-down';
import chevronUp from '@iconify/icons-mdi/chevron-up';
import { InlineIcon } from '@iconify/react';
import { Theme } from '@material-ui/core';
import Box from '@material-ui/core/Box';
import Button, { ButtonProps } from '@material-ui/core/Button';
import Collapse from '@material-ui/core/Collapse';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
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
import { Field, Form, Formik, FormikHelpers, FormikProps } from 'formik';
import { TextField } from 'formik-material-ui';
import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import * as Yup from 'yup';
import API from '../../api/API';
import { Application, Group, Instance } from '../../api/apiDataTypes';
import {
  ERROR_STATUS_CODE,
  getErrorAndFlags,
  getInstanceStatus,
  makeLocaleTime,
  prepareErrorMessage
} from '../../utils/helpers';
import ChannelItem from '../Channels/Item';
import { CardFeatureLabel, CardLabel } from '../Common/Card';
import Empty from '../Common/EmptyContent';
import ListHeader from '../Common/ListHeader';
import Loader from '../Common/Loader';
import MoreMenu from '../Common/MoreMenu';
import makeStatusDefs from './StatusDefs';

const useDetailsStyles = makeStyles((theme: Theme) => ({
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
  },
}));

const useRowStyles = makeStyles((theme: Theme) => ({
  statusExplanation: {
    padding: theme.spacing(2),
  },
  root: {
    '& .MuiTableCell-root': {
      padding: '0.5rem',
    },
  },
}));

const useStatusStyles = makeStyles({
  statusButton: {
    textTransform: 'unset',
    verticalAlign: 'bottom',
  },
  // Align text with icon
  statusText: {
    display: 'inline',
    verticalAlign: 'bottom',
    lineHeight: '30px',
  },
});

interface StatusLabelProps {
  status: Instance['statusInfo'];
  activated?: boolean;
  onClick?: ButtonProps['onClick'];
}

function StatusLabel(props: StatusLabelProps) {
  const classes = useStatusStyles();
  const statusDefs = makeStatusDefs(useTheme());
  const { status, activated } = props;
  const { icon = null, label = 'Unknown', color } = (status && statusDefs[status.type]) || {};
  const iconSize = '22px';

  return (
    <span>
      {/* If there is no onClick passed to it, then we're not a button */}
      {props.onClick ? (
        <Button size="small" onClick={props.onClick} className={classes.statusButton}>
          <Box
            bgcolor={status?.bgColor}
            color={status?.textColor}
            p={0.8}
            display="inline-block"
            mr={1}
          >
            {label}
          </Box>
          <InlineIcon
            icon={activated ? chevronUp : chevronDown}
            height="25"
            width="25"
            color="#808080"
          />
        </Button>
      ) : (
        <Typography className={classes.statusText}>{label}</Typography>
      )}
    </span>
  );
}

interface StatusEvent {
  status: number;
  version: string;
  error_code?: number;
  created_ts: string | Date | number;
}

interface StatusRow {
  entry: StatusEvent;
}

function StatusRow(props: StatusRow) {
  const classes = useRowStyles();
  const { entry } = props;
  const time = makeLocaleTime(entry.created_ts);
  const status = getInstanceStatus(entry.status, entry.version);
  const [collapsed, setCollapsed] = React.useState(true);
  let extendedErrorLabel = '';
  const errorCode = entry.error_code;
  if (entry.status === ERROR_STATUS_CODE && !!errorCode) {
    const [errorMessages, flags] = getErrorAndFlags(errorCode);
    extendedErrorLabel = prepareErrorMessage(errorMessages, flags);
  }
  function onStatusClick() {
    setCollapsed(!collapsed);
  }

  return (
    <React.Fragment>
      <TableRow className={classes.root}>
        <TableCell>
          <StatusLabel onClick={onStatusClick} activated={!collapsed} status={status} />
        </TableCell>
        <TableCell>{entry.version}</TableCell>
        <TableCell>{time}</TableCell>
      </TableRow>
      <TableRow>
        <TableCell padding="none" colSpan={3}>
          <Collapse in={!collapsed}>
            <Typography className={classes.statusExplanation}>
              {status.explanation}
              {extendedErrorLabel && (
                <>
                  {':'}
                  <Box>{extendedErrorLabel}</Box>
                </>
              )}
            </Typography>
          </Collapse>
        </TableCell>
      </TableRow>
    </React.Fragment>
  );
}

function EventTable(props: {events: StatusEvent[]}) {
  return props.events.length === 0 ? (
    <Empty>No events to report for this instance yet.</Empty>
  ) : (
    <Table>
      <TableHead>
        <TableRow>
          <TableCell>Status</TableCell>
          <TableCell>Version</TableCell>
          <TableCell>Time</TableCell>
        </TableRow>
      </TableHead>
      <TableBody>
        {props.events.map((entry, i) => (
          <StatusRow key={`status-row-${i}`} entry={entry} />
        ))}
      </TableBody>
    </Table>
  );
}

interface EditDialogProps {
  show: boolean;
  onHide: (instance?: Instance) => void;
  instance: Instance;
}

interface FormValues {
  name: string;
}

function EditDialog(props: EditDialogProps) {
  const { show, onHide, instance } = props;

  function handleClose() {
    onHide();
  }

  function handleSubmit(values: FormValues, actions: FormikHelpers<FormValues>) {
    actions.setSubmitting(true);

    API.updateInstance(instance.id, values.name)
      .then(updatedInstance => {
        actions.setSubmitting(false);
        onHide(updatedInstance);
      })
      .catch(err => {
        actions.setSubmitting(false);
        actions.setStatus({
          statusMessage:
            err && err.message ? `Something went wrong: ${err.message}` : 'Something went wrong…',
        });
      });
  }

  function renderForm({ status, isSubmitting }: FormikProps<FormValues>) {
    return (
      <Form data-testid="instance-edit-form">
        <DialogContent>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Field
            name="name"
            component={TextField}
            margin="dense"
            label="Name"
            type="text"
            helperText="Leave empty for displaying the instance ID"
            fullWidth
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            Save
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    name: Yup.string().max(256, 'Must enter a valid name (less than 256 characters)'),
  });

  return (
    <Dialog open={show} onClose={handleClose} aria-labelledby="form-dialog-title" fullWidth>
      <DialogTitle>Edit Instance</DialogTitle>
      <Formik
        initialValues={{
          name: instance.alias || instance.id,
        }}
        onSubmit={handleSubmit}
        // validationSchema={validation}
        render={renderForm}
      />
    </Dialog>
  );
}

interface DetailsViewProps {
  application: Application;
  group: Group;
  instance: Instance;
  onInstanceUpdated: () => void;
}

function DetailsView(props: DetailsViewProps) {
  const classes = useDetailsStyles();
  const theme = useTheme<Theme>();
  const { application, group, instance, onInstanceUpdated } = props;
  const [eventHistory, setEventHistory] = React.useState<StatusEvent[] | null>(null);
  const [showEdit, setShowEdit] = React.useState(false);

  const hasAlias = !!instance.alias;

  React.useEffect(() => {
    API.getInstanceStatusHistory(application.id, group.id, instance.id)
      .then(statusHistory => {
        setEventHistory(statusHistory || []);
      })
      .catch(() => {
        setEventHistory([]);
      });
  }, [instance]);

  function updateInstance() {
    setShowEdit(true);
  }

  function onEditHide(newInstance?: Instance) {
    setShowEdit(false);
    if (newInstance !== null) {
      onInstanceUpdated();
    }
  }

  return (
    <>
      <ListHeader title="Instance Information" />
      <Paper>
        <Box p={2}>
          <Grid container spacing={1}>
            <Grid item xs={12}>
              <Grid container justify="space-between">
                <Grid item>
                  <Box fontWeight={700} fontSize={30} color={theme.palette.greyShadeColor}>
                    {instance.alias || instance.id}
                  </Box>
                </Grid>
                <Grid item>
                  <MoreMenu
                    options={[
                      {
                        label: 'Rename',
                        action: updateInstance,
                      },
                    ]}
                  />
                </Grid>
              </Grid>
            </Grid>
            <Grid item md>
              <Box mt={2}>
                {application && group && instance && (
                  <Grid container>
                    <Grid item xs={12} container>
                      <Grid item container>
                        {hasAlias && (
                          <Grid item xs={12}>
                            <CardFeatureLabel>ID</CardFeatureLabel>&nbsp;
                            <Box mt={1} mb={1}>
                              <CardLabel>{instance.id}</CardLabel>
                            </Box>
                          </Grid>
                        )}
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
                      <Grid item xs={12}>
                        <Divider className={classes.divider} />
                      </Grid>

                      <Grid item xs={12} container>
                        <Grid item xs={6}>
                          <CardFeatureLabel>Status</CardFeatureLabel>
                          <Box mt={1}>
                            <StatusLabel status={instance.statusInfo} />
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

                      <Grid item xs={12}>
                        <Divider className={classes.divider} />
                      </Grid>
                      <Grid item xs={12} container>
                        <Grid item xs={6}>
                          <CardFeatureLabel>Application</CardFeatureLabel>
                          <Box mt={1}>
                            <Link
                              className={classes.link}
                              to={`/apps/${application.id}`}
                              component={RouterLink}
                            >
                              {application.name}
                            </Link>
                          </Box>
                        </Grid>
                        <Grid item xs={6}>
                          <CardFeatureLabel>Group</CardFeatureLabel>
                          <Box mt={1}>
                            <Link
                              className={classes.link}
                              to={`/apps/${application.id}/groups/${group.id}`}
                              component={RouterLink}
                            >
                              {group.name}
                            </Link>
                          </Box>
                        </Grid>
                      </Grid>

                      <Grid item xs={12}>
                        <Box mt={2}>
                          <CardFeatureLabel>Channel</CardFeatureLabel>&nbsp;
                          {group.channel ? (
                            <ChannelItem channel={group.channel} />
                          ) : (
                            <CardLabel>None</CardLabel>
                          )}
                        </Box>
                      </Grid>
                    </Grid>
                  </Grid>
                )}
              </Box>
            </Grid>
            <Box width="1%">
              <Divider orientation="vertical" variant="fullWidth" />
            </Box>
            <Grid item md>
              <Box mt={2} fontSize={18} fontWeight={700} color={theme.palette.greyShadeColor}>
                Event Timeline
                {eventHistory ? (
                  <Box padding="1em">
                    <div className={classes.timelineContainer}>
                      <EventTable events={eventHistory} />
                    </div>
                  </Box>
                ) : (
                  <Loader />
                )}
              </Box>
            </Grid>
          </Grid>
        </Box>
      </Paper>
      <EditDialog show={showEdit} onHide={onEditHide} instance={instance} />
    </>
  );
}

export default DetailsView;
