import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemSecondaryAction from '@mui/material/ListItemSecondaryAction';
import ListItemText from '@mui/material/ListItemText';
import MenuItem from '@mui/material/MenuItem';
import MuiSelect, { SelectChangeEvent } from '@mui/material/Select';
import { styled } from '@mui/material/styles';
import MuiTextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import { Field, Form, Formik, FormikHelpers, FormikProps } from 'formik';
import { TextField } from 'formik-mui';
import React from 'react';
import { useTranslation } from 'react-i18next';
import * as Yup from 'yup';

import API from '../../api/API';
import { Channel, Package } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { ARCHES, cleanSemverVersion } from '../../utils/helpers';
import AutoCompletePicker from '../common/AutoCompletePicker';
import ColorPicker from '../common/ColorPicker';

const PREFIX = 'ChannelEdit';

const classes = {
  nameField: `${PREFIX}-nameField`,
};

const StyledDialog = styled(Dialog)({
  [`& .${classes.nameField}`]: {
    width: '15rem',
  },
});

const PackagesPerPage = 15;

interface ChannelFormValues {
  name: string;
  package?: string;
}

interface FormStatus {
  statusMessage?: string;
}

export interface ChannelEditProps {
  data: {
    applicationID: string;
    channel?: Channel;
    packages?: Package[];
  };
  create?: boolean;
  show: boolean;
  onHide: () => void;
}

function ChannelEdit(props: ChannelEditProps) {
  const { t } = useTranslation();

  // Helper to format packages for AutoCompletePicker suggestions
  const formatPackageSuggestions = React.useCallback(
    (packages: Package[]) => {
      return packages.map(pkg => {
        const date = new Date(pkg.created_ts);
        return {
          primary: pkg.version,
          secondary: t('channels|created', { date }),
        };
      });
    },
    [t]
  );

  // Helper to find package by version
  const findPackageByVersion = React.useCallback((packages: Package[], version: string) => {
    return packages.find(pkg => pkg.version === version);
  }, []);
  const defaultColor = '';
  const [channelColor, setChannelColor] = React.useState(defaultColor);
  const [packages, setPackages] = React.useState<{ total: number; packages: Package[] }>({
    packages: [],
    total: 0,
  });
  const defaultArch = 1;
  const [arch, setArch] = React.useState(defaultArch);
  const isCreation = Boolean(props.create);
  const inputSearchTimeout = 250; // ms
  const [packageSearchTerm, setPackageSearchTerm] = React.useState<string>('');
  const [searchPage, setSearchPage] = React.useState<number>(0);
  const [floorPackages, setFloorPackages] = React.useState<Package[]>([]);
  const [loadingFloors, setLoadingFloors] = React.useState(false);
  const [showAddFloorDialog, setShowAddFloorDialog] = React.useState(false);
  const [selectedFloorPackage, setSelectedFloorPackage] = React.useState<Package | null>(null);
  const [floorReason, setFloorReason] = React.useState<string>('');

  // Memoize filtered packages to avoid repeated filtering in render
  const packagesForArch = React.useMemo(
    () =>
      packages.packages.filter((pkg: Package) => {
        // Filter by architecture
        if (pkg.arch !== arch) return false;

        // Filter out packages that have this channel blacklisted
        // (only for existing channels, not for new channel creation)
        if (!isCreation && props.data.channel?.id && pkg.channels_blacklist) {
          if (pkg.channels_blacklist.includes(props.data.channel.id)) {
            return false;
          }
        }

        return true;
      }),
    [packages.packages, arch, isCreation, props.data.channel?.id]
  );

  const availableFloorPackages = React.useMemo(
    () => packagesForArch.filter(pkg => !floorPackages.some(fp => fp.id === pkg.id)),
    [packagesForArch, floorPackages]
  );

  React.useEffect(() => {
    setArch(props.data.channel ? props.data.channel.arch : defaultArch);
    setChannelColor(props.data.channel ? props.data.channel.color : defaultColor);
  }, [props.data]);

  // Fetch floor packages when showing an existing channel
  React.useEffect(() => {
    if (!isCreation && props.data.channel?.id && props.show) {
      setLoadingFloors(true);
      API.getChannelFloors(props.data.channel.id)
        .then(({ packages }) => {
          setFloorPackages(packages || []);
        })
        .catch(e => {
          console.error('Failed to get floor packages: ', e);
          setFloorPackages([]);
        })
        .finally(() => {
          setLoadingFloors(false);
        });
    } else if (!props.show) {
      // Reset all state when dialog closes
      setFloorPackages([]);
      setSelectedFloorPackage(null);
      setFloorReason('');
      setShowAddFloorDialog(false);
      // Also clear search state
      setPackageSearchTerm('');
      setSearchPage(0);
    }
  }, [props.data.channel?.id, isCreation, props.show]);

  async function deleteFloorPackage(packageID: string) {
    if (!props.data.channel?.id) return;

    if (window.confirm(t('channels|confirm_delete_floor'))) {
      try {
        await applicationsStore().deleteChannelFloor(props.data.channel.id, packageID);
        setFloorPackages(prev => prev.filter(p => p.id !== packageID));
      } catch (err) {
        console.error('Failed to delete floor package:', err);
        alert(t('channels|failed_to_delete_floor'));
      }
    }
  }

  async function handleAddFloor() {
    if (!selectedFloorPackage || !selectedFloorPackage.id) return;
    if (!props.data.channel?.id) return;

    try {
      await applicationsStore().setChannelFloor(
        props.data.channel.id,
        selectedFloorPackage.id,
        floorReason || undefined
      );

      const packageWithReason = { ...selectedFloorPackage, floor_reason: floorReason };
      setFloorPackages(prev => [...prev, packageWithReason]);

      setShowAddFloorDialog(false);
      setSelectedFloorPackage(null);
      setFloorReason('');
    } catch (err) {
      console.error('Failed to add floor package:', err);
      alert(t('channels|failed_to_add_floor'));
    }
  }

  function handleSubmit(values: ChannelFormValues, actions: FormikHelpers<ChannelFormValues>) {
    const data: {
      name: string;
      arch: number;
      color: string;
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
      data['id'] = props.data.channel!.id;
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
    // Don't fetch if dialog is not shown
    if (!props.show) return;

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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [packageSearchTerm, props.show]);

  React.useEffect(() => {
    // Don't fetch if dialog is not shown
    if (!props.show) return;

    if (searchPage === 0) {
      // This is handled by the term search.
      return;
    }

    fetchPackages(packageSearchTerm, searchPage);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchPage, props.show]);

  function loadMorePackages() {
    if ((searchPage + 1) * PackagesPerPage < packages.total) {
      setSearchPage(page => page + 1);
    }
  }

  function renderForm({
    values,
    status,
    setFieldValue,
    isSubmitting,
  }: FormikProps<ChannelFormValues> & {
    status?: FormStatus;
  }) {
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
            <Grid>
              <ColorPicker color={channelColor} onColorPicked={color => setChannelColor(color.hex)}>
                <>{values.name ? values.name[0] : ''}</>
              </ColorPicker>
            </Grid>
            <Grid container alignItems="flex-start" spacing={2}>
              <Grid className={classes.nameField}>
                <Field
                  name="name"
                  component={TextField}
                  variant="standard"
                  margin="dense"
                  label={t('frequent|name')}
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
            <FormHelperText>{t('channels|cannot_be_changed')}</FormHelperText>
          </FormControl>
          <Field
            type="text"
            name="package"
            label={t('frequent|package')}
            select
            margin="dense"
            component={AutoCompletePicker}
            helperText={t('channels|showing_only_for_architecture', {
              arch: ARCHES[arch],
            })}
            fullWidth
            onSelect={(packageVersion: string) => {
              const selectedPackage = findPackageByVersion(packagesForArch, packageVersion);
              if (selectedPackage) {
                setFieldValue('package', selectedPackage.id);
              }
            }}
            suggestions={formatPackageSuggestions(packagesForArch)}
            placeholder={t('channels|pick_package')}
            pickerPlaceholder={t('channels|search_package_prompt')}
            dialogTitle={t('channels|choose_package')}
            defaultValue={
              props.data.channel && props.data.channel.package
                ? props.data.channel.package.version
                : ''
            }
            onValueChanged={(term?: string | null) => {
              setPackageSearchTerm(term || '');
              setSearchPage(0);
            }}
            onBottomScrolled={loadMorePackages}
          />
          {!isCreation && (
            <Box mt={2}>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Typography variant="subtitle2" color="textSecondary">
                  {t('channels|floor_packages')} ({floorPackages.length})
                </Typography>
                <Button
                  size="small"
                  startIcon={<AddIcon />}
                  onClick={() => setShowAddFloorDialog(true)}
                  aria-label="Add floor package to this channel"
                >
                  {t('channels|add_floor')}
                </Button>
              </Box>
              <Typography variant="caption" color="textSecondary" display="block" gutterBottom>
                {t('channels|floor_packages_help')}
              </Typography>
              {loadingFloors ? (
                <CircularProgress size={20} />
              ) : floorPackages.length > 0 ? (
                <List dense>
                  {floorPackages.map(pkg => (
                    <ListItem key={pkg.id}>
                      <ListItemText
                        primary={cleanSemverVersion(pkg.version)}
                        secondary={pkg.floor_reason || t('channels|no_reason_specified')}
                      />
                      <ListItemSecondaryAction>
                        <IconButton
                          edge="end"
                          aria-label={`Remove floor package ${cleanSemverVersion(pkg.version)}`}
                          onClick={() => pkg.id && deleteFloorPackage(pkg.id)}
                          size="small"
                        >
                          <DeleteIcon fontSize="small" />
                        </IconButton>
                      </ListItemSecondaryAction>
                    </ListItem>
                  ))}
                </List>
              ) : (
                <Typography variant="body2" color="textSecondary">
                  {t('channels|no_floor_packages')}
                </Typography>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => props.onHide()} color="primary">
            {t('frequent|cancel')}
          </Button>
          <Button type="submit" disabled={isSubmitting} color="primary">
            {isCreation ? t('frequent|add_lower') : t('frequent|save')}
          </Button>
        </DialogActions>
      </Form>
    );
  }

  const maxChars = 50;
  const validation = Yup.object().shape({
    name: Yup.string()
      .max(maxChars, t('common|max_length_error', { number: maxChars }))
      .required('Required'),
  });

  let initialValues: Partial<ChannelFormValues> = {
    name: '',
    package: '',
  };
  if (!isCreation) {
    initialValues = {
      name: props.data.channel?.name || '',
      package: props.data.channel?.package_id || '',
    };
  }

  return (
    <>
      <StyledDialog
        open={props.show}
        onClose={() => {
          // Clear search state when closing dialog
          setPackageSearchTerm('');
          setSearchPage(0);
          props.onHide();
        }}
        aria-labelledby="form-dialog-title"
      >
        <DialogTitle>
          {isCreation ? t('channels|add_new_channel') : t('channels|edit_channel')}
        </DialogTitle>
        <Formik<ChannelFormValues>
          initialValues={initialValues as ChannelFormValues}
          onSubmit={handleSubmit}
          validationSchema={validation}
        >
          {renderForm}
        </Formik>
      </StyledDialog>

      {/* Add Floor Package Dialog */}
      <Dialog
        open={showAddFloorDialog}
        onClose={() => {
          setShowAddFloorDialog(false);
          setSelectedFloorPackage(null);
          setFloorReason('');
        }}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>{t('channels|add_floor_package')}</DialogTitle>
        <DialogContent>
          <AutoCompletePicker
            label={t('frequent|package')}
            defaultValue=""
            onSelect={(packageVersion: string) => {
              const selectedPackage = findPackageByVersion(availableFloorPackages, packageVersion);
              if (selectedPackage) {
                setSelectedFloorPackage(selectedPackage);
              }
            }}
            suggestions={formatPackageSuggestions(availableFloorPackages)}
            placeholder={t('channels|pick_package')}
            pickerPlaceholder={t('channels|search_package_prompt')}
            dialogTitle={t('channels|choose_floor_package')}
            onValueChanged={(term?: string | null) => {
              setPackageSearchTerm(term || '');
              setSearchPage(0);
            }}
            onBottomScrolled={loadMorePackages}
          />
          <FormHelperText>
            {t('channels|showing_only_for_architecture', { arch: ARCHES[arch] })}
          </FormHelperText>
          <MuiTextField
            fullWidth
            margin="dense"
            variant="standard"
            label={t('channels|floor_reason')}
            value={floorReason}
            onChange={e => setFloorReason(e.target.value)}
            aria-label="Floor reason explanation"
            aria-describedby="floor-reason-dialog-helper"
            helperText={
              <span id="floor-reason-dialog-helper">{t('channels|floor_reason_help')}</span>
            }
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowAddFloorDialog(false)}>{t('frequent|cancel')}</Button>
          <Button onClick={handleAddFloor} color="primary" disabled={!selectedFloorPackage}>
            {t('frequent|add')}
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

export default ChannelEdit;
