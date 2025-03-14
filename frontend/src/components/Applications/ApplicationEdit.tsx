import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import MenuItem from '@mui/material/MenuItem';
import { Field, Form, Formik } from 'formik';
import { TextField } from 'formik-mui';
import { useTranslation } from 'react-i18next';
import * as Yup from 'yup';
import { Application } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';

export interface ApplicationEditProps {
  create?: any;
  data: any;
  show: boolean;
  onHide: () => void;
}

export default function ApplicationEdit(props: ApplicationEditProps) {
  const isCreation = Boolean(props.create);
  const { t } = useTranslation();

  function handleSubmit(
    values: { [key: string]: any },
    actions: {
      [key: string]: any;
    }
  ) {
    var data = {
      name: values.name,
      description: values.description,
      product_id: values.product_id,
    };

    let appFunctionCall;
    if (isCreation) {
      if (values.appToClone === 'none') {
        values.appToClone = '';
      }
      appFunctionCall = applicationsStore().createApplication(data, values.appToClone);
    } else {
      appFunctionCall = applicationsStore().updateApplication(props.data.id, data);
    }

    appFunctionCall
      .then(() => {
        actions.setSubmitting(false);
        props.onHide();
      })
      .catch(() => {
        actions.setSubmitting(false);
        actions.setStatus({
          statusMessage: t('common|something_wrong'),
        });
      });
  }

  function handleClose() {
    props.onHide();
  }

  //@ts-ignore
  function renderForm({ status, isSubmitting }) {
    return (
      <Form data-testid="app-edit-form">
        <DialogContent>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Field
            name="name"
            component={TextField}
            variant="standard"
            margin="dense"
            label={t('frequent|Name')}
            type="text"
            fullWidth
            required
          />
          <Field
            name="product_id"
            component={TextField}
            variant="standard"
            margin="dense"
            label={t('frequent|Product ID')}
            type="text"
            fullWidth
            helperText={t('applications|example_app_id')}
          />
          <Field
            name="description"
            component={TextField}
            variant="standard"
            margin="dense"
            label={t('frequent|Description')}
            type="text"
            fullWidth
          />
          {isCreation && (
            <Field
              type="text"
              name="appToClone"
              variant="standard"
              label={t('applications|groups_channels')}
              select
              helperText={t('applications|clone_channels_groups_from_another_app')}
              margin="normal"
              component={TextField}
              InputLabelProps={{
                shrink: true,
              }}
            >
              <MenuItem value="none" key="none">
                {t('applications|do_not_copy')}
              </MenuItem>
              {props.data.applications &&
                props.data.applications.map((application: Application, i: number) => {
                  return (
                    <MenuItem value={application.id} key={'app_' + i}>
                      {application.name}
                    </MenuItem>
                  );
                })}
            </Field>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            {t('frequent|Cancel')}
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            {isCreation ? t('frequent|Add') : t('frequent|Update')}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const maxNameChars = 50;
  const maxDescChars = 155;
  const validation = Yup.object().shape({
    name: Yup.string()
      .max(maxNameChars, t('common|Must be less than x characters', { number: maxNameChars }))
      .required('Required'),
    product_id: Yup.string()
      // This regex matches an ID that matches
      // * At least two segments.
      // * All characters must be alphanumeric, or a dash.
      // Each segment must start with a letter.
      // Each segment must not end with a dash.
      .matches(
        /^[a-zA-Z]+([a-zA-Z0-9\-]*[a-zA-Z0-9])*(\.[a-zA-Z]+([a-zA-Z0-9\-]*[a-zA-Z0-9])*)+$/,
        t('common|Must be a reverse domain ID like io.example.MyApp')
      )
      .nullable(),
    description: Yup.string().max(
      maxDescChars,
      t('common|Must be less than x characters', { number: maxDescChars })
    ),
  });

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title">
      <DialogTitle>
        {isCreation ? t('applications|add_application') : t('applications|update_application')}
      </DialogTitle>
      <Formik
        initialValues={{
          name: props.data.name || '',
          description: props.data.description || '',
          product_id: props.data.product_id || '',
          appToClone: 'none',
        }}
        onSubmit={handleSubmit}
        validationSchema={validation}
      >
        {/* @todo add better types for renderForm */}
        {/* @ts-ignore */}
        {renderForm}
      </Formik>
    </Dialog>
  );
}
