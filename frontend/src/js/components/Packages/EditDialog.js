import _ from "underscore"
import {REGEX_SEMVER} from "../../constants/regex"
import Button from '@material-ui/core/Button';
import Checkbox from '@material-ui/core/Checkbox';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';
import FormHelperText from '@material-ui/core/FormHelperText';
import Grid from '@material-ui/core/Grid';
import InputLabel from '@material-ui/core/InputLabel';
import ListItemText from '@material-ui/core/ListItemText';
import MenuItem from '@material-ui/core/MenuItem';
import MuiSelect from '@material-ui/core/Select';
import PropTypes from 'prop-types';
import React from 'react';
import * as Yup from 'yup';
import { applicationsStore } from '../../stores/Stores';
import { Formik, Form, Field } from 'formik';
import { Select, TextField } from 'formik-material-ui';

function EditDialog(props) {
  const [flatcarType, otherType] = [1, 4];
  const [packageType, setPackageType] =
    React.useState(props.data.channel ? props.data.channel.type : flatcarType);
  const isCreation = Boolean(props.create);

  function getFlatcarActionHash() {
    return props.data.channel.flatcar_action ? props.data.channel.flatcar_action.sha256 : '';
  }

  function isFlatcarType(_type) {
    return _type === flatcarType;
  }

  function getChannelsNames(channelIds) {
    let channels = props.data.channels.filter((channel) => {
      return channelIds.includes(channel.id);
    });
    return channels.map((channelObj) => { return channelObj.name });
  }

  function handlePackageTypeChange(event) {
    setPackageType(event.target.value);
  }

  function handleSubmit(values, actions) {

    let data = {
      filename: values.filename,
      description: values.description,
      url: values.url,
      version: values.version,
      type: packageType,
      size: values.size.toString(),
      hash: values.hash,
      application_id: isCreation ? props.data.appID : props.data.channel.application_id,
      channels_blacklist: values.channelsBlacklist ? values.channelsBlacklist : []
    }

    console.log(packageType);
    console.log(packageType === otherType);

    if (isFlatcarType(packageType)) {
      data.flatcar_action = {sha256: values.flatcarHash}
    }

    let packageFunctionCall;
    if (isCreation) {
      packageFunctionCall = applicationsStore.createPackage(data);
    } else {
      data['id'] = props.data.channel.id;
      packageFunctionCall = applicationsStore.updatePackage(data);
    }

    packageFunctionCall.
      done(() => {
        props.onHide()
        actions.setSubmitting(false);
      }).
      fail(() => {
        actions.setSubmitting(false);
        actions.setStatus({statusMessage: 'Something went wrong, or the version you are trying to add already exists. Check the form or try again later...'});
      })
  }

  function handleClose() {
    props.onHide();
  }

  function renderForm({values, status, isSubmitting}) {
    const channels = props.data.channels ? props.data.channels : [];
    return (
      <Form>
        <DialogContent>
          {status && status.statusMessage &&
          <DialogContentText color="error">
            {status.statusMessage}
          </DialogContentText>
          }
          <FormControl margin="dense">
            <InputLabel>Type</InputLabel>
            <MuiSelect
              value={packageType}
              onChange={handlePackageTypeChange}>
              <MenuItem value={otherType} key="other">Other</MenuItem>
              <MenuItem value={flatcarType} key="flatcar">Flatcar</MenuItem>
            </MuiSelect>
          </FormControl>
          <Field
            name="url"
            component={TextField}
            margin="dense"
            label="URL"
            type="url"
            required={true}
            fullWidth
          />
          <Field
            name="filename"
            component={TextField}
            margin="dense"
            label="Filename"
            type="text"
            required={true}
            fullWidth
          />
          <Field
            name="description"
            component={TextField}
            margin="dense"
            label="Description"
            type="text"
            required={true}
            fullWidth
          />
          <Grid container
            justify="space-between"
            spacing={4}>
            <Grid item xs={6}>
              <Field
                name="version"
                component={TextField}
                margin="dense"
                label="Version"
                type="text"
                required={true}
                helperText="Use SemVer format (1.0.1)"
                fullWidth
              />
            </Grid>
            <Grid item xs={6}>
              <Field
                name="size"
                component={TextField}
                margin="dense"
                label="Size"
                type="number"
                required={true}
                helperText="In bytes"
                fullWidth
              />
            </Grid>
          </Grid>
          <Field
            name="hash"
            component={TextField}
            margin="dense"
            label="Hash"
            type="text"
            required={true}
            helperText="Tip: cat update.gz | openssl dgst -sha1 -binary | base64"
            fullWidth
          />
          {isFlatcarType(packageType) &&
          <Field
            name="flatcarHash"
            component={TextField}
            margin="dense"
            label="Flatcar Action SHA256"
            type="text"
            required={true}
            helperText="Tip: cat update.gz | openssl dgst -sha256 -binary | base64"
            fullWidth
          />
          }
          <FormControl margin="dense" fullWidth>
            <InputLabel>Channels Blacklist</InputLabel>
            <Field
              name="channelsBlacklist"
              component={Select}
              multiple
              renderValue={selected => getChannelsNames(selected).join(' / ')}
            >
              {channels.map((packageItem) => {
                let label = packageItem.name;
                let isDisabled = !isCreation && packageItem.package &&
                  props.data.channel.version === packageItem.package.version;

                return (
                  <MenuItem value={packageItem.id} disabled={isDisabled} key={packageItem.id}>
                    <Checkbox checked={values.channelsBlacklist.indexOf(packageItem.id) > -1} />
                    <ListItemText primary={label} secondary={ isDisabled ? "channel pointing to this package" : null } />
                  </MenuItem>
                );
              })
              }
            </Field>
            <FormHelperText>Blacklisted channels cannot point to this package</FormHelperText>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">Cancel</Button>
          <Button type="submit" disabled={isSubmitting} color="primary">{ isCreation ? "Add" : "Save" }</Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    url: Yup.string().url(),
    filename: Yup.string()
      .max(100, 'Must enter a valid filename (less than 100 characters)')
      .required('Required'),
    // @todo: Validate whether the version already exists so we can provide
    // better feedback.
    version: Yup.string()
      .matches(REGEX_SEMVER, 'Enter a valid semver (1.0.1)')
      .required('Required'),
    size: Yup.number()
      .integer('Must be an integer number')
      .positive('Must be a positive number')
      .required('Required'),
    hash: Yup.string()
      .max(64, 'Must be a valid hash (less than 64 characters)')
      .required('Required'),
  });

  let initialValues = {channelsBlacklist: []};
  if (!isCreation) {
    validation['flatcarHash'] = Yup.string()
      .max(64, 'Must be a valid hash (less than 64 characters)')
      .required('Required');

    initialValues = {url: props.data.channel.url,
                     filename: props.data.channel.filename,
                     description: props.data.channel.description,
                     version: props.data.channel.version,
                     size: props.data.channel.size,
                     hash: props.data.channel.hash,
                     channelsBlacklist: props.data.channel.channels_blacklist ? props.data.channel.channels_blacklist : [],
                    };

    if (isFlatcarType(packageType)) {
      initialValues['flatcarHash'] = getFlatcarActionHash();
    }
  }

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title">
      <DialogTitle>{ isCreation ? "Add Package" : "Edit Package" }</DialogTitle>
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
