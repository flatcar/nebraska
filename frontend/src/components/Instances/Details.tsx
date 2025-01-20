import chevronDown from '@iconify/icons-mdi/chevron-down';
import chevronUp from '@iconify/icons-mdi/chevron-up';
import { InlineIcon } from '@iconify/react';
import { Theme } from '@mui/material';
import Box from '@mui/material/Box';
import Button, { ButtonProps } from '@mui/material/Button';
import Collapse from '@mui/material/Collapse';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import Link from '@mui/material/Link';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import { makeStyles, useTheme } from '@mui/styles';
import { Field, Form, Formik, FormikHelpers, FormikProps } from 'formik';
import { TextField } from 'formik-mui';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link as RouterLink } from 'react-router-dom';
import * as Yup from 'yup';
import API from '../../api/API';
import { Application, Group, Instance, InstanceStatusHistory } from '../../api/apiDataTypes';
import { makeLocaleTime } from '../../i18n/dateTime';
import {
  ERROR_STATUS_CODE,
  getErrorAndFlags,
  getInstanceStatus,
  prepareErrorMessage,
} from '../../utils/helpers';
import ChannelItem from '../Channels/ChannelItem';
import { CardFeatureLabel, CardLabel } from '../common/Card';
import Empty from '../common/EmptyContent';
import ListHeader from '../common/ListHeader';
import Loader from '../common/Loader/Loader';
import MoreMenu from '../common/MoreMenu/MoreMenu';
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
    color: '#1b5c91',
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
  const { t } = useTranslation();

  const { status, activated } = props;
  const { label = t('frequent|Unknown') } = (status && statusDefs[status.type]) || {};
  let safeLabel: React.ReactNode;

  if (label !== null && typeof label === 'object') {
    safeLabel = label.toString();
  } else {
    safeLabel = label;
  }

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
            {safeLabel}
          </Box>
          <InlineIcon
            icon={activated ? chevronUp : chevronDown}
            height="25"
            width="25"
            color="#808080"
          />
        </Button>
      ) : (
        <Typography className={classes.statusText}>{safeLabel}</Typography>
      )}
    </span>
  );
}

interface StatusRow {
  entry: InstanceStatusHistory;
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

function EventTable(props: { events: InstanceStatusHistory[] }) {
  const { t } = useTranslation();

  return props.events.length === 0 ? (
    <Empty>{t('instances|No events to report for this instance yet.')}</Empty>
  ) : (
    <Table>
      <TableHead>
        <TableRow>
          <TableCell>{t('instances|Status')}</TableCell>
          <TableCell>{t('instances|Version')}</TableCell>
          <TableCell>{t('instances|Time')}</TableCell>
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
  const { t } = useTranslation();

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
            err && err.message
              ? t('instances|Something went wrong: {{message}}', { message: err.message })
              : t('instances|Something went wrong…'),
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
            variant="standard"
            margin="dense"
            label={t('instances|Name')}
            type="text"
            helperText={t('instances|Leave empty for displaying the instance ID')}
            fullWidth
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            {t('frequent|Cancel')}
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            {t('frequent|Save')}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    name: Yup.string().max(256, t('instances|Must enter a valid name (less than 256 characters)')),
  });

  return (
    <Dialog open={show} onClose={handleClose} aria-labelledby="form-dialog-title" fullWidth>
      <DialogTitle>{t('instances|Edit Instance')}</DialogTitle>
      <Formik
        initialValues={{
          name: instance.alias || instance.id,
        }}
        onSubmit={handleSubmit}
        validationSchema={validation}
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
  const [eventHistory, setEventHistory] = React.useState<InstanceStatusHistory[] | null>(null);
  const [showEdit, setShowEdit] = React.useState(false);
  const { t } = useTranslation();

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
      <ListHeader title={t('instances|Instance Information')} />
      <Paper>
        <Box p={2}>
          <Grid container spacing={1}>
            <Grid item xs={12}>
              <Grid container justifyContent="space-between">
                <Grid item>
                  <Box fontWeight={700} fontSize={30} color={theme.palette.greyShadeColor}>
                    {instance.alias || instance.id}
                  </Box>
                </Grid>
                <Grid item>
                  <MoreMenu
                    options={[
                      {
                        label: t('instances|Rename'),
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
                            <CardFeatureLabel>{t('instances|ID')}</CardFeatureLabel>&nbsp;
                            <Box mt={1} mb={1}>
                              <CardLabel>{instance.id}</CardLabel>
                            </Box>
                          </Grid>
                        )}
                        <Grid item xs={6}>
                          <CardFeatureLabel>{t('instances|IP')}</CardFeatureLabel>
                          <Box mt={1}>
                            <CardLabel>{instance.ip}</CardLabel>
                          </Box>
                        </Grid>
                        <Grid item xs={6}>
                          <CardFeatureLabel>{t('instances|Version')}</CardFeatureLabel>
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
                          <CardFeatureLabel>{t('instances|Status')}</CardFeatureLabel>
                          <Box mt={1}>
                            <StatusLabel status={instance.statusInfo} />
                          </Box>
                        </Grid>
                        <Grid item xs={6}>
                          <CardFeatureLabel>{t('instances|Last Update Check')}</CardFeatureLabel>
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
                          <CardFeatureLabel>{t('instances|Application')}</CardFeatureLabel>
                          <Box mt={1}>
                            <Link
                              className={classes.link}
                              to={`/apps/${application.id}`}
                              component={RouterLink}
                              underline="hover"
                            >
                              {application.name}
                            </Link>
                          </Box>
                        </Grid>
                        <Grid item xs={6}>
                          <CardFeatureLabel>{t('instances|Group')}</CardFeatureLabel>
                          <Box mt={1}>
                            <Link
                              className={classes.link}
                              to={`/apps/${application.id}/groups/${group.id}`}
                              component={RouterLink}
                              underline="hover"
                            >
                              {group.name}
                            </Link>
                          </Box>
                        </Grid>
                      </Grid>

                      <Grid item xs={12}>
                        <Box mt={2}>
                          <CardFeatureLabel>{t('instances|Channel')}</CardFeatureLabel>&nbsp;
                          {group.channel ? (
                            <ChannelItem channel={group.channel} />
                          ) : (
                            <CardLabel>{t('frequent|None')}</CardLabel>
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
                {t('instances|Event Timeline')}
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
