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
import { makeStyles } from '@material-ui/core/styles';
import { Field, Form, Formik } from 'formik';
import { Select, TextField } from 'formik-material-ui';
import React from 'react';
import { useTranslation } from 'react-i18next';
import * as Yup from 'yup';
import { Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES } from '../../utils/helpers';
import { REGEX_SEMVER } from '../../utils/regex';

const useStyles = makeStyles({
  topSelect: {
    width: '10rem',
  },
  textarea: {
    resize: 'both',
  },
});

function EditDialog(props: { data: any; show: boolean; create?: boolean; onHide: () => void }) {
  const classes = useStyles();
  const [flatcarType, otherType] = [1, 4];
  const [packageType, setPackageType] = React.useState(
    props.data.channel ? props.data.channel.type : flatcarType
  );
  const [arch, setArch] = React.useState(props.data.channel ? props.data.channel.arch : 1);
  const { t } = useTranslation();
  const isCreation = Boolean(props.create);

  function getFlatcarActionHash() {
    return props.data.channel.flatcar_action ? props.data.channel.flatcar_action.sha256 : '';
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

  function handlePackageTypeChange(event: React.ChangeEvent<{ name?: string; value: unknown }>) {
    setPackageType(event.target.value as number);
  }

  function handleArchChange(event: React.ChangeEvent<{ name?: string; value: unknown }>) {
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
      size: values.size?.toString(),
      hash: values.hash,
      metadata_type: values.metadata_type,
      metadata_content: values.metadata_content,
      application_id:
        isCreation && props.data.appID ? props.data.appID : props.data.channel.application_id,
      channels_blacklist: values.channelsBlacklist ? values.channelsBlacklist : [],
    };

    if (isFlatcarType(packageType)) {
      data.flatcar_action = { sha256: values.flatcarHash };
    }

    let pkgFunc: Promise<void>;
    if (isCreation) {
      pkgFunc = applicationsStore.createPackage(data);
    } else {
      pkgFunc = applicationsStore.updatePackage({ ...data, id: props.data.channel.id });
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
  function renderForm({ values, status, isSubmitting }) {
    const channels = props.data.channels ? props.data.channels : [];
    return (
      <Form data-testid="package-edit-form">
        <DialogContent>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Grid container justify="space-between">
            <Grid item>
              <FormControl margin="dense" className={classes.topSelect}>
                <InputLabel>Type</InputLabel>
                <MuiSelect value={packageType} onChange={handlePackageTypeChange}>
                  <MenuItem value={otherType} key="other">
                    {t('packages|Other')}
                  </MenuItem>
                  <MenuItem value={flatcarType} key="flatcar">
                    {t('packages|Flatcar')}
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
                <InputLabel>{t('packages|Architecture')}</InputLabel>
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
                <FormHelperText>{t('packages|Cannot be changed once created.')}</FormHelperText>
              </FormControl>
            </Grid>
          </Grid>
          <Field
            name="url"
            component={TextField}
            margin="dense"
            label={t('packages|URL')}
            type="url"
            required={isFlatcarType(packageType)}
            fullWidth
          />
          <Field
            name="filename"
            component={TextField}
            margin="dense"
            label={t('packages|Filename')}
            type="text"
            required={isFlatcarType(packageType)}
            fullWidth
          />
          <Field
            name="description"
            component={TextField}
            margin="dense"
            label={t('packages|Description')}
            type="text"
            required
            fullWidth
          />
          <Grid container justify="space-between" spacing={4}>
            <Grid item xs={6}>
              <Field
                name="version"
                component={TextField}
                margin="dense"
                label={t('packages|Version')}
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
                margin="dense"
                label={t('packages|Size')}
                type="number"
                required={isFlatcarType(packageType)}
                helperText={t('packages|In bytes')}
                fullWidth
              />
            </Grid>
          </Grid>
          <Field
            name="hash"
            component={TextField}
            margin="dense"
            label={t('packages|Hash')}
            type="text"
            required={isFlatcarType(packageType)}
            helperText={t('packages|Tip: cat update.gz | openssl dgst -sha1 -binary | base64')}
            fullWidth
          />
          {isFlatcarType(packageType) ? (
            <Field
              name="flatcarHash"
              component={TextField}
              margin="dense"
              label={t('packages|Flatcar Action SHA256')}
              type="text"
              required
              helperText={t('packages|Tip: cat update.gz | openssl dgst -sha256 -binary | base64')}
              fullWidth
            />
          ) : (
            <>
              <Field
                name="metadata_type"
                component={TextField}
                margin="dense"
                label={t('packages|Metadata Content Type')}
                type="text"
                helperText={t('packages|E.g. text/json')}
                fullWidth
              />
              <Field
                name="metadata_content"
                component={TextField}
                margin="dense"
                label={t('packages|Metadata Content')}
                type="texta"
                multiline
                rows={4}
                fullWidth
                variant="outlined"
                inputProps={{ className: classes.textarea }}
              />
            </>
          )}

          <FormControl margin="dense" fullWidth>
            <InputLabel>Channels Blacklist</InputLabel>
            <Field
              name="channelsBlacklist"
              component={Select}
              multiple
              renderValue={(selected: string[]) => getChannelsNames(selected).join(' / ')}
            >
              {channels
                .filter((channelItem: Channel) => channelItem.arch === arch)
                .map((packageItem: Channel) => {
                  const label = packageItem.name;
                  const isDisabled =
                    !isCreation &&
                    packageItem.package &&
                    props.data.channel.version === packageItem.package.version;

                  return (
                    <MenuItem value={packageItem.id} disabled={isDisabled} key={packageItem.id}>
                      <Checkbox checked={values.channelsBlacklist.indexOf(packageItem.id) > -1} />
                      <ListItemText
                        primary={label}
                        secondary={
                          isDisabled ? t('packages|channel pointing to this package') : null
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
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose} color="primary">
            {t('frequent|Cancel')}
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            {isCreation ? t('frequent|Add') : t('frequent|Save')}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const validation: {
    [key: string]: any;
  } = Yup.object().shape({
    url: Yup.string().url(),
    filename: Yup.string().max(
      100,
      t('packages|Must enter a valid filename (less than 100 characters)')
    ),
    // @todo: Validate whether the version already exists so we can provide
    // better feedback.
    version: Yup.string()
      .matches(REGEX_SEMVER, t('packages|Enter a valid semver (1.0.1)'))
      .required(t('frequent|Required')),
    size: Yup.number()
      .integer(t('packages|Must be an integer number'))
      .positive(t('packages|Must be a positive number')),
    hash: Yup.string().max(64, t('packages|Must be a valid hash (less than 64 characters)')),
  });

  if (isFlatcarType(packageType)) {
    ['filename', 'size', 'hash', 'url'].forEach(schema => {
      validation.fields[schema] = validation.fields[schema].required(t('frequent|Required'));
    });
  }

  let initialValues: { [key: string]: any } = { channelsBlacklist: [] };
  if (!isCreation) {
    validation['flatcarHash'] = Yup.string()
      .max(64, t('packages|Must be a valid hash (less than 64 characters)'))
      .required(t('frequent|Required'));

    initialValues = {
      url: props.data.channel.url,
      filename: props.data.channel.filename,
      description: props.data.channel.description,
      version: props.data.channel.version,
      size: props.data.channel.size,
      hash: props.data.channel.hash,
      channelsBlacklist: props.data.channel.channels_blacklist
        ? props.data.channel.channels_blacklist
        : [],
      metadata_type: props.data.channel.metadata_type,
      metadata_content: props.data.channel.metadata_content,
    };

    if (isFlatcarType(packageType)) {
      initialValues['flatcarHash'] = getFlatcarActionHash();
    }
  }

  return (
    <Dialog open={props.show} onClose={handleClose} aria-labelledby="form-dialog-title">
      <DialogTitle>
        {isCreation ? t('packages|Add Package') : t('packages|Edit Package')}
      </DialogTitle>
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

export default EditDialog;
