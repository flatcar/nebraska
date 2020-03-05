import { ListItemText } from '@material-ui/core';
import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';
import FormControlLabel from '@material-ui/core/FormControlLabel';
import FormHelperText from '@material-ui/core/FormHelperText';
import Grid from '@material-ui/core/Grid';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import { Field, Form, Formik } from 'formik';
import { Select, Switch, TextField } from 'formik-material-ui';
import PropTypes from 'prop-types';
import React from 'react';
import * as Yup from 'yup';
import { ARCHES } from '../../constants/helpers';
import { applicationsStore } from '../../stores/Stores';
import TimezonePicker, { DEFAULT_TIMEZONE } from '../Common/TimezonePicker';

function EditDialog(props) {
  const isCreation = Boolean(props.create);

  function handleSubmit(values, actions) {
    let updatesPeriodPolicy = values.updatesPeriodRange.toString() + ' '
      + values.updatesPeriodUnit;
    let updatesTimeoutPolicy = values.updatesTimeout.toString() + ' '
      + values.updatesTimeoutUnit;

    let data = {
      name: values.name,
      description: values.description,
      policy_updates_enabled: values.updatesEnabled,
      policy_safe_mode: values.safeMode,
      policy_office_hours: values.onlyOfficeHours,
      policy_max_updates_per_period: values.maxUpdates,
      policy_period_interval: updatesPeriodPolicy,
      policy_update_timeout: updatesTimeoutPolicy,
    }

    if (values.channel)
      data['channel_id'] = values.channel;

    if (values.timezone)
      data['policy_timezone'] = values.timezone;

    let packageFunctionCall;
    if (isCreation) {
      data['application_id'] = props.data.appID;
      packageFunctionCall = applicationsStore.createGroup(data);
    } else {
      data['id'] = props.data.group.id;
      packageFunctionCall = applicationsStore.updateGroup(data);
    }

    packageFunctionCall.
      done(() => {
        props.onHide()
        actions.setSubmitting(false);
      }).
      fail(() => {
        actions.setSubmitting(false);
        actions.setStatus({statusMessage: 'Something went wrong. Check the form or try again later...'});
      })
  }

  function handleClose() {
    props.onHide();
  }

  function renderForm({values, status, setFieldValue, isSubmitting}) {
    const channels = props.data.channels ? props.data.channels : [];

    return (
      <Form>
        <DialogContent>
          {status && status.statusMessage &&
          <DialogContentText color="error">
            {status.statusMessage}
          </DialogContentText>
          }
          <Field
            name="name"
            component={TextField}
            margin="dense"
            label="Name"
            required
            fullWidth
          />
          <Field
            name="description"
            component={TextField}
            margin="dense"
            label="Description"
            fullWidth
          />
          <FormControl margin="dense" fullWidth>
            <InputLabel shrink>Channel</InputLabel>
            <Field
              name="channel"
              component={Select}
              displayEmpty
            >
              <MenuItem value="" key="">
                None yet
              </MenuItem>
              {channels.map((channelItem) =>
                <MenuItem value={channelItem.id} key={channelItem.id}>
                      <ListItemText primary={channelItem.name}
                      secondary={ARCHES[channelItem.arch]} />
                </MenuItem>)
              })}
            </Field>
          </FormControl>
          <Grid container
            justify="space-between"
            spacing={4}>
            <Grid item xs={6}>
              <FormControlLabel
                label="Updates enabled"
                control={
                  <Field
                    name="updatesEnabled"
                    component={Switch}
                    color="primary"
                  />
                }
              />
            </Grid>
            <Grid item xs={6}>
              <FormControlLabel
                label="Safe mode"
                control={
                  <Field
                    name="safeMode"
                    component={Switch}
                    color="primary"
                  />
                }
              />
              <FormHelperText>Only update 1 instance at a time, and stop if an update fails.</FormHelperText>
            </Grid>
          </Grid>
          <Field
            component={TimezonePicker}
            name="timezone"
            value={values.timezone}
            onSelect={timezone => {
              setFieldValue('timezone', timezone);
            }}
          />
          <FormControl fullWidth>
            <FormControlLabel
              label="Only office hours"
              control={
                <Field
                  name="onlyOfficeHours"
                  component={Switch}
                  color="primary"
                />
              }
            />
            <FormHelperText>Only update from 9am to 5pm.</FormHelperText>
          </FormControl>
          <Grid container
            justify="space-between"
            spacing={4}>
            <Grid item xs={6}>
              <Field
                name="maxUpdates"
                label="Max number of updates"
                component={TextField}
                margin="dense"
                type="number"
                fullWidth
              />
            </Grid>
            <Grid item xs={3}>
              <Field
                name="updatesPeriodRange"
                label="Period range"
                component={TextField}
                margin="dense"
                type="number"
                fullWidth
              />
            </Grid>
            <Grid item xs={3}>
              <Field
                name="updatesPeriodUnit"
                label="Period unit"
                component={TextField}
                margin="dense"
                select
                fullWidth
              >
              {['hours', 'minutes', 'days'].map((unit) => {
                return (<MenuItem value={unit} key={unit}>
                          {unit}
                        </MenuItem>);
              })
              }
              </Field>
            </Grid>
          </Grid>
          <Grid container
            spacing={1}>
            <Grid item xs={6}>
              <Field
                name="updatesTimeout"
                label="Updates timeout range"
                component={TextField}
                margin="dense"
                type="number"
              />
            </Grid>
            <Grid item xs={6}>
              <Field
                name="updatesTimeoutUnit"
                label="Timeout unit"
                component={TextField}
                margin="dense"
                select
                fullWidth
              >
              {['hours', 'minutes', 'days'].map((unit) => {
                return (<MenuItem value={unit} key={unit}>
                          {unit}
                        </MenuItem>);
              })
              }
              </Field>
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">Cancel</Button>
          <Button type="submit" disabled={isSubmitting} color="primary">{ isCreation ? "Add" : "Save" }</Button>
        </DialogActions>
      </Form>
    );
  }

  function positiveNum() {
    return Yup.number()
      .positive()
      .min(1, 'Must be greather than or equal to 1')
      .required('Required');
  }

  function maxCharacters(maxChars, required=false) {
    let validation = Yup.string()
      .max(maxChars, `Must be less than $maxChars characters`);

    if (required)
      validation = validation.required('Required');

    return validation;
  }

  const validation = Yup.object().shape({
    name: maxCharacters(50, true),
    description: maxCharacters(250),
    maxUpdates: positiveNum(),
    updatesPeriodRange: positiveNum(),
    updatesTimeout: positiveNum(),
  });

  let initialValues = {}

  if (isCreation) {
    initialValues = {
      maxUpdates: 1,
      updatesPeriodRange: 1,
      updatesPeriodUnit: 'hours',
      updatesTimeout: 1,
      updatesTimeoutUnit: 'days',
      channel: '',
      timezone: DEFAULT_TIMEZONE,
    }
  } else {
    let group = props.data.group;
    let [currentUpdatesPeriodRange, currentUpdatesPeriodUnit] =
      group.policy_period_interval.split(' ');
    let [currentupdatesTimeout, currentUpdatesTimeoutUnit] =
      group.policy_update_timeout.split(' ');

    initialValues = {
      name: group.name,
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
      <DialogTitle>{ isCreation ? "Add Group" : "Edit Group" }</DialogTitle>
        <Formik
          initialValues={initialValues}
          onSubmit={handleSubmit}
          validationSchema={validation}
          render={renderForm}
        />
    </Dialog>
  );
}

EditDialog.propTypes = {
  data: PropTypes.object,
  show: PropTypes.bool,
  create: PropTypes.bool,
}

export default EditDialog;
