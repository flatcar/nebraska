import { Box, Divider, useTheme } from '@mui/material';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import makeStyles from '@mui/styles/makeStyles';
import { Form, Formik } from 'formik';
import React from 'react';
import { useTranslation } from 'react-i18next';
import { useParams } from 'react-router-dom';
import * as Yup from 'yup';
import { Group } from '../../../api/apiDataTypes';
import { applicationsStore } from '../../../stores/Stores';
import { DEFAULT_TIMEZONE } from '../../common/TimezonePicker';
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

export interface GroupEditDialogProps {
  create?: boolean;
  data: {
    [key: string]: any;
  };
  onHide: () => void;
  show: boolean;
}

export default function GroupEditDialog(props: GroupEditDialogProps) {
  const isCreation = Boolean(props.create);
  const classes = useStyles();
  const [groupEditActiveTab, setGroupEditActiveTab] = React.useState(0);
  const { t } = useTranslation();
  const theme = useTheme();
  const { appID } = useParams<{ appID: string }>();

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
      policy_max_updates_per_period: parseInt(values.maxUpdates),
      policy_period_interval: updatesPeriodPolicy,
      policy_update_timeout: updatesTimeoutPolicy,
    };

    if (values.channel) data['channel_id'] = values.channel;

    if (values.timezone) data['policy_timezone'] = values.timezone;

    let packageFunctionCall;
    data['application_id'] = appID;
    if (isCreation) {
      packageFunctionCall = applicationsStore().createGroup(data as Group);
    } else {
      data['id'] = props.data.group.id;
      packageFunctionCall = applicationsStore().updateGroup(data as Group);
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
      .min(1, t('common|Must be greather than or equal to x', { number: 1 }))
      .required('Required');
  }

  function maxCharacters(maxChars: number, required = false) {
    let validation = Yup.string().max(
      maxChars,
      t('common|Must be less than x characters', { number: maxChars })
    );

    if (required) validation = validation.required('Required');

    return validation;
  }

  const validation = Yup.object().shape({
    name: maxCharacters(50, true).required(),
    track: maxCharacters(256),
    description: maxCharacters(250),
    maxUpdates: positiveNum(),
    updatesPeriodRange: positiveNum(),
    updatesTimeout: positiveNum(),
  });

  let initialValues: {
    [key: string]: any;
  } = {
    appID,
  };

  if (isCreation) {
    initialValues = {
      appID: appID,
      maxUpdates: 1,
      updatesPeriodRange: 1,
      updatesPeriodUnit: 'hours',
      updatesTimeout: 1,
      updatesTimeoutUnit: 'days',
      channel: '',
      description: '',
      timezone: DEFAULT_TIMEZONE,
      updatesEnabled: false,
      onlyOfficeHours: false,
      safeMode: false,
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
