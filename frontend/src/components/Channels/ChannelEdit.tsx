import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import Grid from '@mui/material/Grid';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import MuiSelect, { SelectChangeEvent } from '@mui/material/Select';
import makeStyles from '@mui/styles/makeStyles';
import { Field, Form, Formik } from 'formik';
import { TextField } from 'formik-mui';
import React from 'react';
import { useTranslation } from 'react-i18next';
import * as Yup from 'yup';
import API from '../../api/API';
import { Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES } from '../../utils/helpers';
import AutoCompletePicker from '../common/AutoCompletePicker';
import ColorPicker from '../common/ColorPicker';

const useStyles = makeStyles({
  nameField: {
    width: '15rem',
  },
});

const PackagesPerPage = 15;

export interface ChannelEditProps {
  data: any;
  create?: boolean;
  show: boolean;
  onHide: () => void;
}

export default function ChannelEdit(props: ChannelEditProps) {
  const classes = useStyles();
  const { t } = useTranslation();
  const defaultColor = '';
  const [channelColor, setChannelColor] = React.useState(defaultColor);
  const [packages, setPackages] = React.useState<{ total: number; packages: Package[] }>({
    packages: [],
    total: props.data.packages ? -1 : 0,
  });
  const defaultArch = 1;
  const [arch, setArch] = React.useState(defaultArch);
  const isCreation = Boolean(props.create);
  const { channel } = props.data;
  const inputSearchTimeout = 250; // ms
  const [packageSearchTerm, setPackageSearchTerm] = React.useState<string>('');
  const [searchPage, setSearchPage] = React.useState<number>(0);

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
      channelFunctionCall = applicationsStore().createChannel(data as Channel);
    } else {
      data['id'] = props.data.channel.id;
      channelFunctionCall = applicationsStore().updateChannel(data as Channel);
    }

    channelFunctionCall
      .then(() => {
        actions.setSubmitting(false);
        props.onHide();
      })
      .catch(() => {
        actions.setSubmitting(false);
        actions.setStatus({
          statusMessage: t(
            'channels|Something went wrong, or a channel with this name and architecture already exists. Check the form or try again later…'
          ),
        });
      });
  }

  function fetchPackages(term: string, page: number) {
    API.getPackages(props.data.applicationID, term.trim() || '', {
      page: (page || 0) + 1,
      perpage: PackagesPerPage,
    })
      .then(({ packages: pkgs, totalCount }) => {
        setPackages(({ packages }) => ({
          packages: (page === 0 ? [] : packages).concat(pkgs || []),
          total: totalCount,
        }));
      })
      .catch(e => {
        console.error('Failed to get packages for Channels/EditDialog: ', e);
        setPackages({ packages: [], total: 0 });
      });
  }

  React.useEffect(() => {
    let timeoutHandler: NodeJS.Timeout;
    function searchOnTimeout(term: string) {
      if (timeoutHandler !== undefined) {
        // Always clear the timeout before eventually starting a new one.
        clearTimeout(timeoutHandler);
      }

      // Use a timeout to avoid searching on every key strike.
      timeoutHandler = setTimeout(() => {
        fetchPackages(term, 0);
      }, inputSearchTimeout);
    }

    setSearchPage(0);
    searchOnTimeout(packageSearchTerm);

    return function cleanup() {
      if (timeoutHandler !== undefined) {
        // Always clear the timeout on unmount.
        clearTimeout(timeoutHandler);
      }
    };
  }, [packageSearchTerm]);

  React.useEffect(() => {
    if (searchPage === 0) {
      // This is handled by the term search.
      return;
    }

    fetchPackages(packageSearchTerm, searchPage);
  }, [searchPage]);

  function loadMorePackages() {
    if ((searchPage + 1) * PackagesPerPage < packages.total) {
      setSearchPage(page => page + 1);
    }
  }

  //@todo add better types
  //@ts-ignore
  function renderForm({ values, status, setFieldValue, isSubmitting }) {
    return (
      <Form data-testid="channel-edit-form">
        <DialogContent>
          {status && status.statusMessage && (
            <DialogContentText color="error">{status.statusMessage}</DialogContentText>
          )}
          <Grid
            container
            spacing={2}
            justifyContent="space-between"
            alignItems="center"
            wrap="nowrap"
          >
            <Grid item>
              <ColorPicker color={channelColor} onColorPicked={color => setChannelColor(color.hex)}>
                {values.name ? values.name[0] : ''}
              </ColorPicker>
            </Grid>
            <Grid item container alignItems="flex-start" spacing={2}>
              <Grid item className={classes.nameField}>
                <Field
                  name="name"
                  component={TextField}
                  variant="standard"
                  margin="dense"
                  label={t('frequent|Name')}
                  InputLabelProps={{ shrink: true }}
                  type="text"
                  required
                  helperText={t(
                    'channels|Can be an existing one as long as the arch is different.'
                  )}
                  fullWidth
                />
              </Grid>
            </Grid>
          </Grid>
          <FormControl margin="dense" disabled={!isCreation} fullWidth>
            <InputLabel variant="standard">Architecture</InputLabel>
            <MuiSelect
              variant="standard"
              value={arch}
              onChange={(event: SelectChangeEvent<number>) => setArch(event.target.value as number)}
            >
              {Object.keys(ARCHES).map((key: string) => {
                const archName = ARCHES[parseInt(key)];
                return (
                  <MenuItem value={parseInt(key)} key={key}>
                    {archName}
                  </MenuItem>
                );
              })}
            </MuiSelect>
            <FormHelperText>{t('channels|Cannot be changed once created.')}</FormHelperText>
          </FormControl>
          <Field
            type="text"
            name="package"
            label={t('frequent|Package')}
            select
            margin="dense"
            component={AutoCompletePicker}
            helperText={t("channels|Showing only for the channel's architecture ({{arch}}).", {
              arch: ARCHES[arch],
            })}
            fullWidth
            onSelect={(packageVersion: string) => {
              const selectedPackage = packages.packages
                .filter((packageItem: Package) => packageItem.arch === arch)
                .filter((packageItem: Package) => packageItem.version === packageVersion);
              if (selectedPackage.length) {
                setFieldValue('package', selectedPackage[0].id);
              }
            }}
            suggestions={packages.packages
              .filter((packageItem: Package) => packageItem.arch === arch)
              .map((packageItem: Package) => {
                const date = new Date(packageItem.created_ts);
                return {
                  primary: packageItem.version,
                  secondary: t('channels|created: {{date, date}}', { date: date }),
                };
              })}
            placeholder={t('channels|Pick a package')}
            pickerPlaceholder={t('channels|Start typing to search a package')}
            data={packages.packages.filter((packageItem: Package) => packageItem.arch === arch)}
            dialogTitle={t('channels|Choose a package')}
            defaultValue={channel && channel.package ? channel.package.version : ''}
            onValueChanged={(term: string | null) => {
              setPackageSearchTerm(term || '');
              setSearchPage(0);
            }}
            onBottomScrolled={loadMorePackages}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => props.onHide()} color="primary">
            {t('frequent|Cancel')}
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            {isCreation ? t('frequent|Add') : t('frequent|Save')}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const validation = Yup.object().shape({
    name: Yup.string()
      .max(50, t('common|Must be less than x characters', { number: 50 }))
      .required('Required'),
  });

  let initialValues = {};
  if (!isCreation) {
    initialValues = {
      name: props.data.channel.name,
      package: props.data.channel.package_id ? props.data.channel.package_id : '',
    };
  }

  return (
    <Dialog open={props.show} onClose={() => props.onHide()} aria-labelledby="form-dialog-title">
      <DialogTitle>
        {isCreation ? t('channels|Add New Channel') : t('channels|Edit Channel')}
      </DialogTitle>
      <Formik initialValues={initialValues} onSubmit={handleSubmit} validationSchema={validation}>
        {/* @todo add better types */}
        {/* @ts-ignore */}
        {renderForm}
      </Formik>
    </Dialog>
  );
}
