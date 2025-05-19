import chevronDown from '@iconify/icons-mdi/chevron-down';
import chevronUp from '@iconify/icons-mdi/chevron-up';
import { InlineIcon } from '@iconify/react';
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
import { styled } from '@mui/material/styles';
import { useTheme } from '@mui/material/styles';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
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

const PREFIX = 'DetailsView';

const classes = {
  timelineContainer: `${PREFIX}-timelineContainer`,
  divider: `${PREFIX}-divider`,
  link: `${PREFIX}-link`,
};

const StyledPaper = styled(Paper)(({ theme }) => ({
  [`& .${classes.timelineContainer}`]: {
    maxHeight: '700px',
    overflow: 'auto',
  },

  [`& .${classes.divider}`]: {
    marginTop: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },

  [`& .${classes.link}`]: {
    fontSize: '1rem',
    color: '#1b5c91',
  },
}));

interface StatusLabelProps {
  status: Instance['statusInfo'];
  activated?: boolean;
  onClick?: ButtonProps['onClick'];
}

function StatusLabel(props: StatusLabelProps) {
  const theme = useTheme();
  const statusDefs = makeStatusDefs(theme);
  const { t } = useTranslation();

  const { status, activated } = props;
  const { label = t('frequent|unknown') } = (status && statusDefs[status.type]) || {};

  return (
    <span>
      {/* If there is no onClick passed to it, then we're not a button */}
      {props.onClick ? (
        <Button
          size="small"
          onClick={props.onClick}
          sx={{
            textTransform: 'unset',
            verticalAlign: 'bottom',
          }}
        >
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
        <Typography
          sx={{
            display: 'inline',
            verticalAlign: 'bottom',
            lineHeight: '30px',
          }}
        >
          {label}
        </Typography>
      )}
    </span>
  );
}

interface StatusRow {
  entry: InstanceStatusHistory;
}

function StatusRow(props: StatusRow) {
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
      <TableRow
        sx={{
          '& .MuiTableCell-root': {
            padding: '0.5rem',
          },
        }}
      >
        <TableCell>
          <StatusLabel onClick={onStatusClick} activated={!collapsed} status={status} />
        </TableCell>
        <TableCell>{entry.version}</TableCell>
        <TableCell>{time}</TableCell>
      </TableRow>
      <TableRow>
        <TableCell padding="none" colSpan={3}>
          <Collapse in={!collapsed}>
            <Typography
              sx={{
                padding: theme => theme.spacing(2),
              }}
            >
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
    <Empty>{t('instances|no_events_message')}</Empty>
  ) : (
    <Table>
      <TableHead>
        <TableRow>
          <TableCell>{t('instances|status')}</TableCell>
          <TableCell>{t('instances|version')}</TableCell>
          <TableCell>{t('instances|time')}</TableCell>
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
              ? t('instances|max_updates_per_period', { message: err.message })
              : t('instances|error_message'),
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
            label={t('instances|name')}
            type="text"
            helperText={t('instances|leave_empty_for_displaying_the_instance_id')}
            fullWidth
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            {t('frequent|cancel')}
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            {t('frequent|save')}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    name: Yup.string().max(256, t('instances|valid_name_warning')),
  });

  return (
    <Dialog open={show} onClose={handleClose} aria-labelledby="form-dialog-title" fullWidth>
      <DialogTitle>{t('instances|edit_instance')}</DialogTitle>
      <Formik
        initialValues={{
          name: instance.alias || instance.id,
        }}
        onSubmit={handleSubmit}
        validationSchema={validation}
      >
        {renderForm}
      </Formik>
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
  const theme = useTheme();
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
      <ListHeader title={t('instances|instance_information')} />
      <StyledPaper>
        <Box p={2}>
          <Grid container spacing={1}>
            <Grid size={12}>
              <Grid container justifyContent="space-between">
                <Grid>
                  <Box fontWeight={700} fontSize={30} color={theme.palette.greyShadeColor}>
                    {instance.alias || instance.id}
                  </Box>
                </Grid>
                <Grid>
                  <MoreMenu
                    options={[
                      {
                        label: t('instances|rename'),
                        action: updateInstance,
                      },
                    ]}
                  />
                </Grid>
              </Grid>
            </Grid>
            <Grid
              size={{
                md: 'grow',
              }}
            >
              <Box mt={2}>
                {application && group && instance && (
                  <Grid container>
                    <Grid container size={12}>
                      <Grid container sx={{ width: '100%' }}>
                        {hasAlias && (
                          <Grid size={12}>
                            <CardFeatureLabel>{t('instances|id')}</CardFeatureLabel>&nbsp;
                            <Box mt={1} mb={1}>
                              <CardLabel>{instance.id}</CardLabel>
                            </Box>
                          </Grid>
                        )}
                        <Grid size={6}>
                          <CardFeatureLabel>{t('instances|ip')}</CardFeatureLabel>
                          <Box mt={1}>
                            <CardLabel>{instance.ip}</CardLabel>
                          </Box>
                        </Grid>
                        <Grid size={6}>
                          <CardFeatureLabel>{t('instances|version')}</CardFeatureLabel>
                          <Box mt={1}>
                            <CardLabel>{instance.application.version}</CardLabel>
                          </Box>
                        </Grid>
                      </Grid>
                      <Grid size={12}>
                        <Divider className={classes.divider} />
                      </Grid>

                      <Grid container size={12}>
                        <Grid size={6}>
                          <CardFeatureLabel>{t('instances|status')}</CardFeatureLabel>
                          <Box mt={1}>
                            <StatusLabel status={instance.statusInfo} />
                          </Box>
                        </Grid>
                        <Grid size={6}>
                          <CardFeatureLabel>{t('instances|last_update_check')}</CardFeatureLabel>
                          <Box mt={1}>
                            <CardLabel>
                              {makeLocaleTime(instance.application.last_check_for_updates)}
                            </CardLabel>
                          </Box>
                        </Grid>
                      </Grid>

                      <Grid size={12}>
                        <Divider className={classes.divider} />
                      </Grid>
                      <Grid container size={12}>
                        <Grid size={6}>
                          <CardFeatureLabel>{t('instances|application')}</CardFeatureLabel>
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
                        <Grid size={6}>
                          <CardFeatureLabel>{t('instances|group')}</CardFeatureLabel>
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

                      <Grid size={12}>
                        <Box mt={2}>
                          <CardFeatureLabel>{t('instances|channel')}</CardFeatureLabel>&nbsp;
                          {group.channel ? (
                            <ChannelItem channel={group.channel} />
                          ) : (
                            <CardLabel>{t('frequent|none')}</CardLabel>
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
            <Grid
              size={{
                md: 'grow',
              }}
            >
              <Box mt={2} fontSize={18} fontWeight={700} color={theme.palette.greyShadeColor}>
                {t('instances|event_timeline')}
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
      </StyledPaper>
      <EditDialog show={showEdit} onHide={onEditHide} instance={instance} />
    </>
  );
}

export default DetailsView;
