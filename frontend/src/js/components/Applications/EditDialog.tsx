import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import MenuItem from '@material-ui/core/MenuItem';
import { Field, Form, Formik } from 'formik';
import { TextField } from 'formik-material-ui';
import React from 'react';
import * as Yup from 'yup';
import { Application } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';

function EditDialog(props: {
  create?: any;
  data: any;
  show: boolean;
  onHide: () => void;
}) {
  const isCreation = Boolean(props.create);

  function handleSubmit(values: {[key: string]: any}, actions: {
    [key: string]: any;
  }) {
    var data = {
      name: values.name,
      description: values.description
    };

    let appFunctionCall;
    if (isCreation) {
      if (values.appToClone === 'none') {
        values.appToClone = '';
      }
      appFunctionCall = applicationsStore.createApplication(data, values.appToClone);
    } else {
      appFunctionCall = applicationsStore.updateApplication(props.data.id, data);
    }

    appFunctionCall
      .then(() => {
        actions.setSubmitting(false);
        props.onHide();
      })
      .catch(() => {
        actions.setSubmitting(false);
        actions.setStatus({statusMessage: 'Something went wrong. Check the form or try again later...'});
      });
  }

  function handleClose() {
    props.onHide();
  }

  //@ts-ignore
  function renderForm({status, isSubmitting}) {
    return (
      <Form data-testid="app-edit-form">
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
            fullWidth
            required
          />
          <Field
            name="description"
            component={TextField}
            margin="dense"
            label="Description"
            type="text"
            fullWidth
          />
          {isCreation &&
            <Field
              type="text"
              name="appToClone"
              label="Groups/Channels"
              select
              helperText="Clone channels and groups from another other application"
              margin="normal"
              component={TextField}
              InputLabelProps={{
                shrink: true,
              }}
            >
              <MenuItem value="none" key="none">Do not copy</MenuItem>
              {props.data.applications &&
                props.data.applications.map((application: Application, i: number) => {
                  return (
                    <MenuItem value={application.id} key={'app_' + i}>
                      {application.name}
                    </MenuItem>
                  );
                })
              }
            </Field>
          }
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">Cancel</Button>
          <Button type="submit" disabled={isSubmitting} color="primary">{isCreation ? 'Add' : 'Update'}</Button>
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
      <DialogTitle>{isCreation ? 'Add Application' : 'Update Application'}</DialogTitle>
      <Formik
        initialValues={{ name: props.data.name,
                         description: props.data.description }}
        onSubmit={handleSubmit}
        validationSchema={validation}
        //@todo add better types for renderForm
        //@ts-ignore
        render={renderForm}
      />
    </Dialog>
  );
}

export default EditDialog;
