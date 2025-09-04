import Box from '@mui/material/Box';
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
import { styled } from '@mui/material/styles';
import MuiTextField from '@mui/material/TextField';
import { Field, Form, Formik, FormikHelpers, FormikProps } from 'formik';
import { Select, TextField } from 'formik-mui';
import React from 'react';
import { useTranslation } from 'react-i18next';
import * as Yup from 'yup';

import API from '../../api/API';
import { Channel, File, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES } from '../../utils/helpers';
import { REGEX_SEMVER } from '../../utils/regex';
import Tabs from '../common/Tabs';
import FileList from './FileList';

const PREFIX = 'EditDialog';

const classes = {
  topSelect: `${PREFIX}-topSelect`,
  dialog: `${PREFIX}-dialog`,
};

// Package type constants - must match backend pkg/api/packages.go
const PACKAGE_TYPE_FLATCAR = 1; // Flatcar Linux OS update package
const PACKAGE_TYPE_OTHER = 4; // Generic/other package type

const StyledDialog = styled(Dialog)({
  [`& .${classes.topSelect}`]: {
    width: '10rem',
  },
  [`& .${classes.dialog}`]: {
    height: 'calc(100% - 64px)',
  },
});

interface PackageFormValues {
  url: string;
  filename: string;
  description: string;
  version: string;
  size: number | string;
  hash: string;
  flatcarHash?: string;
  channelsBlacklist: string[];
  filesList: File[];
}

interface FormStatus {
  statusMessage?: string;
}

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
  const [packageType, setPackageType] = React.useState(
    props.data.package ? props.data.package.type : PACKAGE_TYPE_FLATCAR
  );
  const [arch, setArch] = React.useState(props.data.package ? props.data.package.arch : 1);
  const { t } = useTranslation();
  const isCreation = Boolean(props.create);
  const [isAddingFiles, setIsAddingFiles] = React.useState(false);
  const [packageFloorChannels, setPackageFloorChannels] = React.useState<string[]>([]);
  const [floorReason, setFloorReason] = React.useState<string>('');
  const [loadingFloors, setLoadingFloors] = React.useState(false);
  const [initialFloorChannels, setInitialFloorChannels] = React.useState<string[]>([]);

  React.useEffect(() => {
    if (!isCreation && props.data.package?.id && props.data.appID && props.show) {
      setLoadingFloors(true);
      setPackageFloorChannels([]);
      setFloorReason('');

      const packageId = props.data.package.id;
      const loadFloorChannels = async () => {
        try {
          const response = await API.getPackageFloorChannels(props.data.appID, packageId);

          if (response.channels && response.channels.length > 0) {
            const channelIds = response.channels.map(item => item.channel.id);
            const firstReason =
              response.channels.find(item => item.floor_reason)?.floor_reason || '';

            setPackageFloorChannels(channelIds);
            setInitialFloorChannels(channelIds);
            setFloorReason(firstReason);
          }
        } catch (e) {
          console.error('Failed to load floor channels:', e);
        } finally {
          setLoadingFloors(false);
        }
      };

      loadFloorChannels();
    } else if (!props.show) {
      setPackageFloorChannels([]);
      setInitialFloorChannels([]);
      setFloorReason('');
      setLoadingFloors(false);
    }
  }, [props.data.package?.id, props.data.appID, props.show, isCreation]);

  function getFlatcarActionHash() {
    return props.data.package.flatcar_action ? props.data.package.flatcar_action.sha256 : '';
  }

  function isFlatcarType(_type: number) {
    return _type === PACKAGE_TYPE_FLATCAR;
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

  function handleSubmit(values: PackageFormValues, actions: FormikHelpers<PackageFormValues>) {
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
      .then(async () => {
        if (!isCreation && props.data.package?.id) {
          const packageId = props.data.package.id;
          const added = packageFloorChannels.filter(id => !initialFloorChannels.includes(id));
          const removed = initialFloorChannels.filter(id => !packageFloorChannels.includes(id));

          if (added.length > 0 || removed.length > 0) {
            try {
              await Promise.all([
                ...added.map(channelId =>
                  applicationsStore().addChannelFloor(
                    channelId,
                    packageId,
                    floorReason || undefined
                  )
                ),
                ...removed.map(channelId =>
                  applicationsStore().deleteChannelFloor(channelId, packageId)
                ),
              ]);
            } catch (err) {
              console.error('Failed to save floor channel changes:', err);
              actions.setStatus({
                statusMessage: t('packages|floor_changes_failed'),
              });
            }
          }
        }

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

  function renderForm({
    values,
    status,
    isSubmitting,
    setValues,
  }: FormikProps<PackageFormValues> & {
    status?: FormStatus;
  }) {
    const channels = props.data.channels ? props.data.channels : [];
    return (
      <Form data-testid="package-edit-form">
        <DialogContent>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Grid container justifyContent="space-between">
            <Grid>
              <FormControl margin="dense" className={classes.topSelect}>
                <InputLabel variant="standard">Type</InputLabel>
                <MuiSelect
                  variant="standard"
                  value={packageType}
                  onChange={handlePackageTypeChange}
                >
                  <MenuItem value={PACKAGE_TYPE_OTHER} key="other">
                    {t('packages|other')}
                  </MenuItem>
                  <MenuItem value={PACKAGE_TYPE_FLATCAR} key="flatcar">
                    {t('packages|flatcar')}
                  </MenuItem>
                </MuiSelect>
              </FormControl>
            </Grid>
            <Grid>
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
                        <Grid size={6}>
                          <Field
                            name="version"
                            component={TextField}
                            variant="standard"
                            margin="dense"
                            label={`${t('packages|version')}`}
                            type="text"
                            required
                            helperText={t('packages|valid_name_warning')}
                            fullWidth
                          />
                        </Grid>
                        <Grid size={6}>
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
                        helperText={t('packages|tip_command', {
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
                          helperText={t('packages|tip_command', {
                            command: 'cat update.gz | openssl dgst -sha256 -binary | base64',
                          })}
                          fullWidth
                        />
                      )}
                      <FormControl margin="dense" fullWidth>
                        <Field
                          name="channelsBlacklist"
                          component={Select}
                          variant="standard"
                          label="Channels Blacklist"
                          multiple
                          renderValue={(selected: string[]) =>
                            getChannelsNames(selected).join(' / ')
                          }
                        >
                          {channels
                            .filter((channelItem: Channel) => channelItem.arch === arch)
                            .map((channelItem: Channel) => {
                              const label = channelItem.name;
                              const isDisabled =
                                (!isCreation &&
                                  channelItem.package &&
                                  props.data.package.version === channelItem.package.version) ||
                                false;
                              return (
                                <MenuItem
                                  value={channelItem.id}
                                  disabled={isDisabled}
                                  key={channelItem.id}
                                >
                                  <Checkbox
                                    checked={values.channelsBlacklist.indexOf(channelItem.id) > -1}
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
                      {!isCreation && (
                        <>
                          <FormControl margin="dense" fullWidth>
                            <InputLabel variant="standard" id="floor-channels-label">
                              Floor Channels
                            </InputLabel>
                            <MuiSelect<string[]>
                              variant="standard"
                              labelId="floor-channels-label"
                              multiple
                              value={packageFloorChannels}
                              aria-label="Select floor channels for this package"
                              aria-describedby="floor-channels-helper"
                              onChange={e => {
                                const newValue = (
                                  typeof e.target.value === 'string'
                                    ? e.target.value.split(',')
                                    : e.target.value
                                ) as string[];

                                setPackageFloorChannels(newValue);

                                if (newValue.length === 0) {
                                  setFloorReason('');
                                }
                              }}
                              renderValue={(selected: string[]) => {
                                if (selected.length === 0) {
                                  return <em>None</em>;
                                }
                                return getChannelsNames(selected).join(', ');
                              }}
                              disabled={loadingFloors}
                            >
                              {channels
                                .filter((channelItem: Channel) => channelItem.arch === arch)
                                .map((channelItem: Channel) => {
                                  const isCurrentTarget =
                                    channelItem.package?.version === props.data.package.version;
                                  return (
                                    <MenuItem key={channelItem.id} value={channelItem.id}>
                                      <Checkbox
                                        checked={packageFloorChannels.includes(channelItem.id)}
                                      />
                                      <ListItemText
                                        primary={channelItem.name}
                                        secondary={
                                          isCurrentTarget
                                            ? t('packages|channel_currently_pointing_here')
                                            : null
                                        }
                                      />
                                    </MenuItem>
                                  );
                                })}
                            </MuiSelect>
                            <FormHelperText id="floor-channels-helper">
                              Floor packages are mandatory intermediate versions for specified
                              channels.
                              <br />
                              Showing only channels with the same architecture ({ARCHES[arch]}).
                            </FormHelperText>
                          </FormControl>
                          <MuiTextField
                            variant="standard"
                            margin="dense"
                            label={t('packages|floor_reason')}
                            type="text"
                            disabled={packageFloorChannels.length === 0}
                            value={floorReason}
                            aria-label="Floor reason explanation"
                            aria-describedby="floor-reason-helper"
                            onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
                              setFloorReason(e.target.value);
                            }}
                            helperText={
                              <span id="floor-reason-helper">
                                {packageFloorChannels.length === 0
                                  ? t('packages|select_floor_channels_first')
                                  : t('packages|floor_reason_help')}
                              </span>
                            }
                            fullWidth
                          />
                        </>
                      )}
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
      .matches(REGEX_SEMVER, t('packages|valid_semver_prompt'))
      .required(t('frequent|required')),
    size: Yup.number()
      .integer(t('common|integer_number_error'))
      .positive(t('common|positive_number_error'))
      .required(t('frequent|required')),
    hash: Yup.string()
      .max(maxHashChars, t('common|valid_hash_error', { number: maxHashChars }))
      .required(t('frequent|required')),
  });

  let initialValues: Partial<PackageFormValues> = {
    url: '',
    filename: '',
    description: '',
    version: '',
    size: '',
    hash: '',
    channelsBlacklist: [],
    filesList: [],
  };

  if (!isCreation) {
    const maxFlatcarHashChars = 64;
    validation['flatcarHash'] = Yup.string()
      .max(maxFlatcarHashChars, t('common|valid_hash_error', { number: maxFlatcarHashChars }))
      .required(t('frequent|required'));

    initialValues = {
      url: props.data.package.url,
      filename: props.data.package.filename || '',
      description: props.data.package.description || '',
      version: props.data.package.version,
      size: props.data.package.size || '',
      hash: props.data.package.hash || '',
      channelsBlacklist: props.data.package.channels_blacklist || [],
      filesList: props.data.package.extra_files || [],
    };

    if (isFlatcarType(packageType)) {
      initialValues.flatcarHash = getFlatcarActionHash();
    }
  }

  return (
    <StyledDialog
      open={props.show}
      onClose={handleClose}
      aria-labelledby="form-dialog-title"
      fullWidth
    >
      <DialogTitle>
        {isCreation ? t('packages|add_package') : t('packages|edit_package')}
      </DialogTitle>
      <Formik<PackageFormValues>
        initialValues={initialValues as PackageFormValues}
        onSubmit={handleSubmit}
        validationSchema={validation}
      >
        {renderForm}
      </Formik>
    </StyledDialog>
  );
}

export default EditDialog;
