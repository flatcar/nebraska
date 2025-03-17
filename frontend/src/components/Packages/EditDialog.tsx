import { Box } from '@mui/material';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import Grid from '@mui/material/Grid';
import InputLabel from '@mui/material/InputLabel';
import ListItemText from '@mui/material/ListItemText';
import MenuItem from '@mui/material/MenuItem';
import MuiSelect, { SelectChangeEvent } from '@mui/material/Select';
import makeStyles from '@mui/styles/makeStyles';
import { Field, Form, Formik } from 'formik';
import { Select, TextField } from 'formik-mui';
import React from 'react';
import { useTranslation } from 'react-i18next';
import * as Yup from 'yup';
import { Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES } from '../../utils/helpers';
import { REGEX_SEMVER } from '../../utils/regex';
import Tabs from '../common/Tabs';
import FileList from './FileList';

const useStyles = makeStyles({
  topSelect: {
    width: '10rem',
  },
  dialog: {
    height: 'calc(100% - 64px)',
  },
});

export interface EditDialogProps {
  create?: boolean;
  data: {
    appID: string;
    channels: Channel[];
    package: Package;
  };
  show: boolean;
  onHide: () => void;
}

function EditDialog(props: EditDialogProps) {
  const classes = useStyles();
  const [flatcarType, otherType] = [1, 4];
  const [packageType, setPackageType] = React.useState(
    props.data.package ? props.data.package.type : flatcarType
  );
  const [arch, setArch] = React.useState(props.data.package ? props.data.package.arch : 1);
  const { t } = useTranslation();
  const isCreation = Boolean(props.create);
  const [isAddingFiles, setIsAddingFiles] = React.useState(false);

  function getFlatcarActionHash() {
    return props.data.package.flatcar_action ? props.data.package.flatcar_action.sha256 : '';
  }

  function isFlatcarType(_type: number) {
    return _type === flatcarType;
  }

  function getChannelsNames(channelIds: string[]) {
    const channels = props.data.channels.filter((channel: Channel) => {
      return channelIds.includes(channel.id);
    });
    return channels.map((channelObj: Channel) => {
      return channelObj.name;
    });
  }

  function handlePackageTypeChange(event: SelectChangeEvent<number>) {
    setPackageType(event.target.value as number);
  }

  function handleArchChange(event: SelectChangeEvent<number>) {
    setArch(event.target.value as number);
  }
  //@todo add better types
  //@ts-ignore
  function handleSubmit(values, actions) {
    const data: Partial<Package> = {
      arch: typeof arch === 'string' ? parseInt(arch) : arch,
      filename: values.filename,
      description: values.description,
      url: values.url,
      version: values.version,
      type: packageType,
      size: values.size.toString(),
      hash: values.hash,
      application_id:
        isCreation && props.data.appID ? props.data.appID : props.data.package.application_id,
      channels_blacklist: values.channelsBlacklist ? values.channelsBlacklist : [],
      extra_files: values.filesList,
    };

    if (isFlatcarType(packageType)) {
      data.flatcar_action = { sha256: values.flatcarHash };
    }

    let pkgFunc: Promise<void>;
    if (isCreation) {
      pkgFunc = applicationsStore().createPackage(data);
    } else {
      pkgFunc = applicationsStore().updatePackage({ ...data, id: props.data.package.id });
    }

    pkgFunc
      .then(() => {
        props.onHide();
        actions.setSubmitting(false);
      })
      .catch(() => {
        actions.setSubmitting(false);
        actions.setStatus({
          statusMessage: t(
            'packages|Something went wrong, or the version you are trying to add already exists for the arch and package type. Check the form or try again later...'
          ),
        });
      });
  }

  function handleClose() {
    props.onHide();
  }

  //@todo add better types
  //@ts-ignore
  function renderForm({ values, status, isSubmitting, setValues }) {
    const channels = props.data.channels ? props.data.channels : [];
    return (
      <Form data-testid="package-edit-form">
        <DialogContent>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Grid container justifyContent="space-between">
            <Grid item>
              <FormControl margin="dense" className={classes.topSelect}>
                <InputLabel variant="standard">Type</InputLabel>
                <MuiSelect
                  variant="standard"
                  value={packageType}
                  onChange={handlePackageTypeChange}
                >
                  <MenuItem value={otherType} key="other">
                    {t('packages|other')}
                  </MenuItem>
                  <MenuItem value={flatcarType} key="flatcar">
                    {t('packages|flatcar')}
                  </MenuItem>
                </MuiSelect>
              </FormControl>
            </Grid>
            <Grid item>
              <FormControl
                margin="dense"
                fullWidth
                className={classes.topSelect}
                disabled={!isCreation}
              >
                <InputLabel variant="standard">{t('packages|architecture')}</InputLabel>
                <MuiSelect variant="standard" value={arch} onChange={handleArchChange}>
                  {Object.keys(ARCHES).map((key: string) => {
                    const archName = ARCHES[parseInt(key)];
                    return (
                      <MenuItem value={parseInt(key)} key={key}>
                        {archName}
                      </MenuItem>
                    );
                  })}
                </MuiSelect>
                <FormHelperText>{t('packages|immutable_warning')}</FormHelperText>
              </FormControl>
            </Grid>
          </Grid>
          <Box maxHeight="calc(100% - 100px)">
            <Tabs
              tabProps={{ centered: true, variant: 'standard' }}
              tabs={[
                {
                  label: t('frequent|main'),
                  component: (
                    <>
                      <Field
                        name="url"
                        component={TextField}
                        variant="standard"
                        margin="dense"
                        label={t('packages|url')}
                        type="url"
                        required
                        fullWidth
                      />
                      <Field
                        name="filename"
                        component={TextField}
                        variant="standard"
                        margin="dense"
                        label={t('packages|filename')}
                        type="text"
                        required
                        fullWidth
                      />
                      <Field
                        name="description"
                        component={TextField}
                        variant="standard"
                        margin="dense"
                        label={t('packages|description')}
                        type="text"
                        required
                        fullWidth
                      />
                      <Grid container justifyContent="space-between" spacing={4}>
                        <Grid item xs={6}>
                          <Field
                            name="version"
                            component={TextField}
                            variant="standard"
                            margin="dense"
                            label={`${t('packages|version')}:`}
                            type="text"
                            required
                            helperText={t('packages|Use SemVer format (1.0.1)')}
                            fullWidth
                          />
                        </Grid>
                        <Grid item xs={6}>
                          <Field
                            name="size"
                            component={TextField}
                            variant="standard"
                            margin="dense"
                            label={t('packages|size')}
                            type="number"
                            required
                            helperText={t('packages|in_bytes')}
                            fullWidth
                          />
                        </Grid>
                      </Grid>
                      <Field
                        name="hash"
                        component={TextField}
                        variant="standard"
                        margin="dense"
                        label={t('packages|hash')}
                        type="text"
                        required
                        helperText={t('packages|Tip: {{command}}', {
                          command: 'cat update.gz | openssl dgst -sha1 -binary | base64',
                        })}
                        fullWidth
                      />
                      {isFlatcarType(packageType) && (
                        <Field
                          name="flatcarHash"
                          component={TextField}
                          variant="standard"
                          margin="dense"
                          label={t('packages|flatcar_action_sha256')}
                          type="text"
                          required
                          helperText={t('packages|Tip: {{command}}', {
                            command: 'cat update.gz | openssl dgst -sha256 -binary | base64',
                          })}
                          fullWidth
                        />
                      )}
                      <FormControl margin="dense" fullWidth>
                        <InputLabel variant="standard">Channels Blacklist</InputLabel>
                        <Field
                          name="channelsBlacklist"
                          component={Select}
                          variant="standard"
                          multiple
                          renderValue={(selected: string[]) =>
                            getChannelsNames(selected).join(' / ')
                          }
                        >
                          {channels
                            .filter((channelItem: Channel) => channelItem.arch === arch)
                            .map((packageItem: Channel) => {
                              const label = packageItem.name;
                              const isDisabled =
                                (!isCreation &&
                                  packageItem.package &&
                                  props.data.package.version === packageItem.package.version) ||
                                false;
                              return (
                                <MenuItem
                                  value={packageItem.id}
                                  disabled={isDisabled}
                                  key={packageItem.id}
                                >
                                  <Checkbox
                                    checked={values.channelsBlacklist.indexOf(packageItem.id) > -1}
                                  />
                                  <ListItemText
                                    primary={label}
                                    secondary={
                                      isDisabled ? t('packages|channel_pointing_to_package') : null
                                    }
                                  />
                                </MenuItem>
                              );
                            })}
                        </Field>
                        <FormHelperText>
                          Blacklisted channels cannot point to this package.
                          <br />
                          Showing only channels with the same architecture ({ARCHES[arch]}).
                        </FormHelperText>
                      </FormControl>
                    </>
                  ),
                },
                {
                  label: t('frequent|extra_files'),
                  component: (
                    <FileList
                      files={values.filesList}
                      onFilesChanged={files => {
                        setValues({ ...values, filesList: files });
                      }}
                      onEditChanged={setIsAddingFiles}
                    />
                  ),
                },
              ]}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            {t('frequent|cancel')}
          </Button>
          <Button type="submit" disabled={isSubmitting || isAddingFiles} color="primary">
            {isCreation ? t('frequent|add_lower') : t('frequent|save')}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const maxFilenameChars = 100;
  const maxHashChars = 64;
  const validation: {
    [key: string]: any;
  } = Yup.object().shape({
    url: Yup.string().url(),
    filename: Yup.string()
      .max(
        maxFilenameChars,
        t('common|valid_filename_error', {
          number: maxFilenameChars,
        })
      )
      .required(t('frequent|required')),
    // @todo: Validate whether the version already exists so we can provide
    // better feedback.
    version: Yup.string()
      .matches(REGEX_SEMVER, t('packages|Enter a valid semver (1.0.1)'))
      .required(t('frequent|required')),
    size: Yup.number()
      .integer(t('common|integer_number_error'))
      .positive(t('common|positive_number_error'))
      .required(t('frequent|required')),
    hash: Yup.string()
      .max(maxHashChars, t('common|valid_hash_error', { number: maxHashChars }))
      .required(t('frequent|required')),
  });

  let initialValues: { [key: string]: any } = { channelsBlacklist: [] };
  if (!isCreation) {
    const maxFlatcarHashChars = 64;
    validation['flatcarHash'] = Yup.string()
      .max(maxFlatcarHashChars, t('common|valid_hash_error', { number: maxFlatcarHashChars }))
      .required(t('frequent|required'));

    initialValues = {
      url: props.data.package.url,
      filename: props.data.package.filename,
      description: props.data.package.description,
      version: props.data.package.version,
      size: props.data.package.size,
      hash: props.data.package.hash,
      channelsBlacklist: props.data.package.channels_blacklist
        ? props.data.package.channels_blacklist
        : [],
      filesList: props.data.package.extra_files,
    };

    if (isFlatcarType(packageType)) {
      initialValues['flatcarHash'] = getFlatcarActionHash();
    }
  }

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title" fullWidth>
      <DialogTitle>
        {isCreation ? t('packages|add_package') : t('packages|edit_package')}
      </DialogTitle>
      <Formik initialValues={initialValues} onSubmit={handleSubmit} validationSchema={validation}>
        {/* @todo add better types for renderForm */}
        {/* @ts-ignore */}
        {renderForm}
      </Formik>
    </Dialog>
  );
}

export default EditDialog;
