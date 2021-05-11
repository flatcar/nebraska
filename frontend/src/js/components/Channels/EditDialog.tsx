import Button from '@material-ui/core/Button';
import Dialog from '@material-ui/core/Dialog';
import DialogActions from '@material-ui/core/DialogActions';
import DialogContent from '@material-ui/core/DialogContent';
import DialogContentText from '@material-ui/core/DialogContentText';
import DialogTitle from '@material-ui/core/DialogTitle';
import FormControl from '@material-ui/core/FormControl';
import FormHelperText from '@material-ui/core/FormHelperText';
import Grid from '@material-ui/core/Grid';
import InputLabel from '@material-ui/core/InputLabel';
import MenuItem from '@material-ui/core/MenuItem';
import MuiSelect from '@material-ui/core/Select';
import { makeStyles } from '@material-ui/core/styles';
import { Field, Form, Formik } from 'formik';
import { TextField } from 'formik-material-ui';
import PropTypes from 'prop-types';
import React from 'react';
import * as Yup from 'yup';
import { Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES } from '../../utils/helpers';
import AutoCompletePicker from '../Common/AutoCompletePicker';
import { ColorPickerButton } from '../Common/ColorPicker';
import ChannelAvatar from './ChannelAvatar';

const useStyles = makeStyles(theme => ({
  nameField: {
    width: '15rem',
  },
}));

function EditDialog(props: { data: any; create?: boolean; show: boolean; onHide: () => void }) {
  const classes = useStyles();
  const defaultColor = '';
  const [channelColor, setChannelColor] = React.useState(defaultColor);
  const defaultArch = 1;
  const [arch, setArch] = React.useState(defaultArch);
  const isCreation = Boolean(props.create);
  const { channel } = props.data;

  React.useEffect(() => {
    setArch(props.data.channel ? props.data.channel.arch : defaultArch);
    setChannelColor(props.data.channel ? props.data.channel.color : defaultColor);
  }, [props.data]);

  function handleSubmit(values: { [key: string]: any }, actions: { [key: string]: any }) {
    const data: {
      name: string;
      arch: number;
      color: any;
      application_id: string;
      package_id?: string;
      id?: string;
    } = {
      name: values.name,
      arch: arch,
      color: channelColor,
      application_id: props.data.applicationID,
    };

    const package_id = values.package;
    if (package_id) {
      data['package_id'] = package_id;
    }

    let channelFunctionCall;
    if (isCreation) {
      channelFunctionCall = applicationsStore.createChannel(data as Channel);
    } else {
      data['id'] = props.data.channel.id;
      channelFunctionCall = applicationsStore.updateChannel(data as Channel);
    }

    channelFunctionCall
      .then(() => {
        actions.setSubmitting(false);
        props.onHide();
      })
      .catch(() => {
        actions.setSubmitting(false);
        actions.setStatus({
          statusMessage:
            'Something went wrong, or a channel with this name and architecture already exists. Check the form or try again later…',
        });
      });
  }

  function handleColorPicked(color: { hex: string }) {
    setChannelColor(color.hex);
  }

  function handleArchChange(event: React.ChangeEvent<{ value: any }>) {
    setArch(event.target.value);
  }

  function handleClose() {
    props.onHide();
  }
  //@todo add better types
  //@ts-ignore
  function renderForm({ values, status, setFieldValue, isSubmitting }) {
    const packages = props.data.packages ? props.data.packages : [];
    return (
      <Form data-testid="channel-edit-form">
        <DialogContent>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Grid container spacing={2} justify="space-between" alignItems="center" wrap="nowrap">
            <Grid item>
              <ColorPickerButton
                color={channelColor}
                onColorPicked={handleColorPicked}
                componentColorProp="color"
              >
                <ChannelAvatar>{values.name ? values.name[0] : ''}</ChannelAvatar>
              </ColorPickerButton>
            </Grid>
            <Grid item container alignItems="flex-start" spacing={2}>
              <Grid item className={classes.nameField}>
                <Field
                  name="name"
                  component={TextField}
                  margin="dense"
                  label="Name"
                  InputLabelProps={{ shrink: true }}
                  autoFocus
                  type="text"
                  required
                  helperText="Can be an existing one as long as the arch is different."
                  fullWidth
                />
              </Grid>
            </Grid>
          </Grid>
          <FormControl margin="dense" disabled={!isCreation} fullWidth>
            <InputLabel>Architecture</InputLabel>
            <MuiSelect value={arch} onChange={handleArchChange}>
              {Object.keys(ARCHES).map((key: string) => {
                const archName = ARCHES[parseInt(key)];
                return (
                  <MenuItem value={parseInt(key)} key={key}>
                    {archName}
                  </MenuItem>
                );
              })}
            </MuiSelect>
            <FormHelperText>Cannot be changed once created.</FormHelperText>
          </FormControl>
          <Field
            type="text"
            name="package"
            label="Package"
            select
            margin="dense"
            component={AutoCompletePicker}
            helperText={`Showing only for the channel's architecture (${ARCHES[arch]}).`}
            fullWidth
            onSelect={(packageVersion: string) => {
              const selectedPackage = packages
                .filter((packageItem: Package) => packageItem.arch === arch)
                .filter((packageItem: Package) => packageItem.version === packageVersion);
              setFieldValue('package', selectedPackage[0].id);
            }}
            getSuggestions={packages
              .filter((packageItem: Package) => packageItem.arch === arch)
              .map((packageItem: Package) => {
                const date = new Date(packageItem.created_ts);
                return {
                  primary: packageItem.version,
                  secondary: `created: ${date.toLocaleString('default', {
                    day: '2-digit',
                    month: '2-digit',
                    year: 'numeric',
                  })}`,
                };
              })}
            placeholder={'Pick a package'}
            pickerPlaceholder={'Start typing to search a package'}
            data={packages.filter((packageItem: Package) => packageItem.arch === arch)}
            dialogTitle={'Choose a package'}
            defaultValue={channel && channel.package ? channel.package.version : ''}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            Cancel
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            {isCreation ? 'Add' : 'Save'}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    name: Yup.string().max(50, 'Must be less than 50 characters').required('Required'),
  });

  let initialValues = {};
  if (!isCreation) {
    initialValues = {
      name: props.data.channel.name,
      package: props.data.channel.package_id ? props.data.channel.package_id : '',
    };
  }

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title">
      <DialogTitle>{isCreation ? 'Add New Channel' : 'Edit Channel'}</DialogTitle>
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

EditDialog.propTypes = {
  data: PropTypes.object,
  show: PropTypes.bool,
  create: PropTypes.bool,
};

export default EditDialog;
