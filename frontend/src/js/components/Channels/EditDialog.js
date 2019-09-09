import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import MenuItem from '@material-ui/core/MenuItem';
import PropTypes from 'prop-types';
import React from 'react';
import moment from 'moment';
import * as Yup from 'yup';
import { applicationsStore } from '../../stores/Stores';
import { Formik, Form, Field } from 'formik';
import { TextField } from 'formik-material-ui';
import { ColorPickerButton } from '../Common/ColorPicker';

function EditDialog(props) {

  const [channelColor, setChannelColor] = React.useState(props.data && props.data.channel ? props.data.channel.color : "#000000");
  const isCreation = Boolean(props.create);

  function handleSubmit(values, actions) {
    let data = {
      name: values.name,
      color: channelColor,
      application_id: props.data.applicationID
    }

    let package_id = values.package;
    if (package_id) {
      data["package_id"] = package_id;
    }

    let channelFunctionCall;
    if (isCreation) {
        channelFunctionCall = applicationsStore.createChannel(data);
    } else {
        data['id'] = props.data.channel.id;
        channelFunctionCall = applicationsStore.updateChannel(data);
    }

    channelFunctionCall.
      done(() => {
        actions.setSubmitting(false);
        props.onHide()
      }).
      fail(() => {
        actions.setSubmitting(false);
        actions.setStatus({statusMessage: 'Something went wrong. Check the form or try again later...'});
      })
  }

  function handleColorPicked(color) {
    setChannelColor(color.hex);
  }

  function handleClose() {
    props.onHide();
  }

  function renderForm({values, status, isSubmitting}) {
    const packages = props.data.packages ? props.data.packages : [];
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
            type="text"
            required={true}
            fullWidth
          />
          <ColorPickerButton color={channelColor} onColorPicked={handleColorPicked}/>
          <Field
            type="text"
            name="package"
            label="Package"
            select
            margin="dense"
            component={TextField}
            fullWidth
          >
            <MenuItem value="" key="none">Nothing yet</MenuItem>
            {packages.map((packageItem, i) =>
            <MenuItem value={packageItem.id} key={"packageItem_" + i}>
              {packageItem.version} &nbsp;&nbsp;(created: {moment.utc(packageItem.created_ts).local().format("DD/MM/YYYY")})
            </MenuItem>
            )}
          </Field>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">Cancel</Button>
          <Button type="submit" disabled={isSubmitting} color="primary">{ isCreation ? "Add" : "Save" }</Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    name: Yup.string()
      .max(50, 'Must be less than 50 characters')
      .required('Required'),
  });

  let initialValues = {};
  if (!isCreation) {
    initialValues = {name: props.data.channel.name,
                     package: props.data.channel.package_id ? props.data.channel.package_id : "",
                    };
  }

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title">
      <DialogTitle>{ isCreation ? "Add New Channel" : "Edit Channel" }</DialogTitle>
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
