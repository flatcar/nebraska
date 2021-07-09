import {
  Box,
  Divider,
  FormLabel,
  makeStyles,
  Switch,
  Tooltip,
  Typography,
  useTheme,
} from '@material-ui/core';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import Tab from '@material-ui/core/Tab';
import Tabs from '@material-ui/core/Tabs';
import { Field, Form, Formik } from 'formik';
import React from 'react';
import { useTranslation } from 'react-i18next';
import * as Yup from 'yup';
import { Channel, Group } from '../../../api/apiDataTypes';
import { applicationsStore } from '../../../stores/Stores';
import { DEFAULT_TIMEZONE } from '../../Common/TimezonePicker';
import GroupDetailsForm from './GroupDetailsForm';
import GroupPolicyForm from './GroupPolicyForm';

const useStyles = makeStyles({
  root: {
    padding: '0.5em 0em',
    overflow: 'hidden',
  },
  indicator: {
    background: '#000',
  },
});

function EditDialog(props: {
  create?: boolean;
  data: {
    [key: string]: any;
  };
  onHide: () => void;
  show: boolean;
}) {
  const isCreation = Boolean(props.create);
  const classes = useStyles();
  const [groupEditActiveTab, setGroupEditActiveTab] = React.useState(0);
  const { t } = useTranslation();
  const theme = useTheme();

  function handleSubmit(values: { [key: string]: any }, actions: { [key: string]: any }) {
    const updatesPeriodPolicy =
      values.updatesPeriodRange.toString() + ' ' + values.updatesPeriodUnit;
    const updatesTimeoutPolicy = values.updatesTimeout.toString() + ' ' + values.updatesTimeoutUnit;

    const data: {
      [key: string]: any;
    } = {
      name: values.name,
      track: values.track,
      description: values.description,
      policy_updates_enabled: values.updatesEnabled,
      policy_safe_mode: values.safeMode,
      policy_office_hours: values.onlyOfficeHours,
      policy_max_updates_per_period: values.maxUpdates,
      policy_period_interval: updatesPeriodPolicy,
      policy_update_timeout: updatesTimeoutPolicy,
    };

    if (values.channel) data['channel_id'] = values.channel;

    if (values.timezone) data['policy_timezone'] = values.timezone;

    let packageFunctionCall;
    if (isCreation) {
      data['application_id'] = props.data.appID;
      packageFunctionCall = applicationsStore.createGroup(data as Group);
    } else {
      data['id'] = props.data.group.id;
      packageFunctionCall = applicationsStore.updateGroup(data as Group);
    }

    packageFunctionCall
      .then(() => {
        props.onHide();
        actions.setSubmitting(false);
      })
      .catch(() => {
        actions.setSubmitting(false);
        actions.setStatus({
          statusMessage: t('groups|Something went wrong. Check the form or try again later...'),
        });
      });
  }

  function TabPanel(props: {
    render: () => React.ReactElement | null;
    index: number;
    value: number;
  }) {
    const { index, value, render } = props;

    return index === value ? render() : null;
  }

  function handleClose() {
    props.onHide();
  }

  function onEditGroupTabChange(event: any, index: any) {
    setGroupEditActiveTab(index);
  }

  //@ts-ignore
  function renderForm({ values, status, setFieldValue, isSubmitting }) {
    const channels = props.data.channels ? props.data.channels : [];

    return (
      <Form data-testid="group-edit-form">
        <DialogContent className={classes.root}>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Box py={1}>
            <Tabs
              value={groupEditActiveTab}
              onChange={onEditGroupTabChange}
              classes={{ indicator: classes.indicator }}
              aria-label="group edit form tabs"
            >
              <Tab label="Details" />
              <Tab label="Policy" />
            </Tabs>
            <Divider />
          </Box>
          <TabPanel
            index={0}
            value={groupEditActiveTab}
            render={() => (
              <GroupDetailsForm channels={channels} values={values} setFieldValue={setFieldValue} />
            )}
          />
          <TabPanel
            index={1}
            value={groupEditActiveTab}
            render={() => <GroupPolicyForm values={values} setFieldValue={setFieldValue} />}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>
            <Box color={theme.palette.greyShadeColor} component="span">
              {t('frequent|Cancel')}
            </Box>
          </Button>
          <Button type="submit" disabled={isSubmitting}>
            <Box component="span" color="#ffff" bgcolor="#000" width="100%" px={1.5} py={1}>
              {isCreation ? t('frequent|Add') : t('frequent|Save')}
            </Box>
          </Button>
        </DialogActions>
      </Form>
    );
  }

  function positiveNum() {
    return Yup.number()
      .positive()
      .min(1, t('groups|Must be greather than or equal to 1'))
      .required('Required');
  }

  function maxCharacters(maxChars: number, required = false) {
    let validation = Yup.string().max(maxChars, t('groups|Must be less than $maxChars characters'));

    if (required) validation = validation.required('Required');

    return validation;
  }

  const validation = Yup.object().shape({
    name: maxCharacters(50, true),
    track: maxCharacters(256),
    description: maxCharacters(250),
    maxUpdates: positiveNum(),
    updatesPeriodRange: positiveNum(),
    updatesTimeout: positiveNum(),
  });

  let initialValues = {};

  if (isCreation) {
    initialValues = {
      maxUpdates: 1,
      updatesPeriodRange: 1,
      updatesPeriodUnit: 'hours',
      updatesTimeout: 1,
      updatesTimeoutUnit: 'days',
      channel: '',
      timezone: DEFAULT_TIMEZONE,
    };
  } else if (!!props.data.group) {
    const group = props.data.group;
    const [currentUpdatesPeriodRange, currentUpdatesPeriodUnit] =
      group.policy_period_interval.split(' ');
    const [currentupdatesTimeout, currentUpdatesTimeoutUnit] =
      group.policy_update_timeout.split(' ');
    initialValues = {
      name: group.name,
      track: group.track,
      description: group.description,
      timezone: group.policy_timezone || DEFAULT_TIMEZONE,
      updatesEnabled: group.policy_updates_enabled,
      onlyOfficeHours: group.policy_office_hours,
      safeMode: group.policy_safe_mode,
      maxUpdates: group.policy_max_updates_per_period,
      channel: group.channel ? group.channel.id : '',
      updatesPeriodRange: currentUpdatesPeriodRange,
      updatesPeriodUnit: currentUpdatesPeriodUnit,
      updatesTimeout: currentupdatesTimeout,
      updatesTimeoutUnit: currentUpdatesTimeoutUnit,
    };
  }

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title">
      <DialogTitle>{isCreation ? t('groups|Add Group') : t('groups|Edit Group')}</DialogTitle>
      <Formik
        initialValues={initialValues}
        onSubmit={handleSubmit}
        validationSchema={validation}
        //@todo add better types
        //@ts-ignore
        render={renderForm}
      />
    </Dialog>
  );
}

export default EditDialog;
