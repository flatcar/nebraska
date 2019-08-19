import PropTypes from 'prop-types';
import { applicationsStore } from "../../stores/Stores"
import React from "react"
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import MenuItem from '@material-ui/core/MenuItem';
import Button from '@material-ui/core/Button';
import { Formik, Form, Field } from 'formik';
import { TextField } from 'formik-material-ui';
import {ColorPickerButton} from '../Common/ColorPicker'
import moment from "moment"

import * as Yup from 'yup';

function ModalAdd(props) {

  const [channelColor, setChannelColor] = React.useState('#00ff00');

  function handleSubmit(values, actions) {
    let data = {
      name: values.name,
      color: channelColor,
      application_id: props.data.applicationID
    }

    let package_id = values.packageChannel;
    if (package_id) {
      data["package_id"] = package_id;
    }

    applicationsStore.createChannel(data).
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

  function renderForm({status, isSubmitting}) {
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
          <ColorPickerButton onColorPicked={handleColorPicked}/>
          <Field
            type="text"
            name="package"
            label="Package"
            select
            margin="dense"
            component={TextField}
            fullWidth
          >
            <MenuItem value="none" key="none">Nothing yet</MenuItem>
            {packages.map((packageItem, i) =>
            <MenuItem value={packageItem.id} key={"packageItem_" + i}>
              {packageItem.version} &nbsp;&nbsp;(created: {moment.utc(packageItem.created_ts).local().format("DD/MM/YYYY")})
            </MenuItem>
            )}
          </Field>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">Cancel</Button>
          <Button type="submit" disabled={isSubmitting} color="primary">Add</Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    name: Yup.string()
      .max(50, 'Must be less than 50 characters')
      .required('Required'),
    description: Yup.string()
      .max(250, 'Must be less than 250 characters'),
  });

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title">
      <DialogTitle>Add New Channel</DialogTitle>
        <Formik
          initialValues={{ name: '',
                           package: 'none' }}
          onSubmit={handleSubmit}
          validationSchema={validation}
          render={renderForm}
        />
    </Dialog>
  );
}

ModalAdd.propTypes = {
  data: PropTypes.object
}

export default ModalAdd
