import {
  Box,
  Divider,
  FormControlLabel,
  FormHelperText,
  FormLabel,
  Grid,
  MenuItem,
  Switch,
  Tooltip,
  Typography,
} from '@material-ui/core';
import HelpOutlineIcon from '@material-ui/icons/HelpOutline';
import { Field } from 'formik';
import { TextField } from 'formik-material-ui';
import { useTranslation } from 'react-i18next';
import TimezonePicker from '../../common/TimezonePicker';

export default function GroupPolicyForm(props: {
  values: { [key: string]: string };
  setFieldValue: (formField: string, value: any) => any;
}) {
  const { t } = useTranslation();
  const { values, setFieldValue } = props;

  return (
    <div style={{ padding: '1rem' }}>
      <Grid container justify="space-between" spacing={4}>
        <Grid item xs={12}>
          <Box mt={1}>
            <FormLabel component="legend">{t('groups|Update')}</FormLabel>
          </Box>
          <Grid container>
            <Grid item xs={6}>
              <FormControlLabel
                label={t('groups|Updates enabled')}
                control={
                  <Field
                    name="updatesEnabled"
                    component={Switch}
                    color="primary"
                    checked={values.updatesEnabled}
                    onChange={(e: any) => {
                      setFieldValue('updatesEnabled', e.target.checked);
                    }}
                  />
                }
              />
            </Grid>
            <Grid item xs={6}>
              <FormControlLabel
                label={t('groups|Safe mode')}
                control={
                  <Field
                    name="safeMode"
                    component={Switch}
                    checked={values.safeMode}
                    color="primary"
                    onChange={(e: any) => {
                      setFieldValue('safeMode', e.target.checked);
                    }}
                  />
                }
              />
              <FormHelperText>
                {t('groups|Only update 1 instance at a time, and stop if an update fails.')}
              </FormHelperText>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
      <Box mt={2}>
        <Divider />
      </Box>
      <Box mt={2}>
        <FormLabel component="legend">{t('groups|Update Limits')}</FormLabel>
      </Box>
      <Box>
        <Grid container>
          <Grid item xs={6}>
            <FormControlLabel
              label={
                <Box display="flex" alignItems="center">
                  <Box pr={0.5}>{t('groups|Only office hours')}</Box>
                  <Box pt={0.1} color="#808080">
                    <Tooltip title={t('groups|Only update from 9am to 5pm.') || ''}>
                      <HelpOutlineIcon fontSize="small" />
                    </Tooltip>
                  </Box>
                </Box>
              }
              control={
                <Box>
                  <Field
                    name="onlyOfficeHours"
                    component={Switch}
                    color="primary"
                    checked={values.onlyOfficeHours}
                    onChange={(e: any) => {
                      setFieldValue('onlyOfficeHours', e.target.checked);
                    }}
                  />
                </Box>
              }
            />
          </Grid>
          <Grid item xs={6}>
            <Field
              component={TimezonePicker}
              name="timezone"
              value={values.timezone}
              onSelect={(timezone: string) => {
                setFieldValue('timezone', timezone);
              }}
            />
          </Grid>
        </Grid>
      </Box>
      <Box my={2}>
        <Grid item xs={12} container spacing={2} justify="space-between" alignItems="center">
          <Grid item xs={5}>
            <Box pl={2}>
              <Field
                name="maxUpdates"
                component={TextField}
                label={t('groups|Max number of updates')}
                margin="dense"
                type="number"
                fullWidth
                minValue={0}
                defaultValue={values.maxUpdates}
                inputProps={{ min: 0 }}
              />
            </Box>
          </Grid>
          <Grid item>
            <Typography color="textSecondary">{t('groups|per')}</Typography>
          </Grid>
          <Grid item xs={3}>
            <Box mt={2}>
              <Field
                name="updatesPeriodRange"
                component={TextField}
                margin="dense"
                type="number"
                fullWidth
                defaultValue={values.updatesPeriodRange}
                inputProps={{ min: 0 }}
              />
            </Box>
          </Grid>
          <Grid item xs={3}>
            <Box mt={2} mr={2}>
              <Field
                name="updatesPeriodUnit"
                component={TextField}
                margin="dense"
                select
                fullWidth
                defaultValue={values.updatesPeriodUnit}
              >
                <MenuItem value={'hours'} key={'hours'}>
                  {t('groups|hours')}
                </MenuItem>
                <MenuItem value={'minutes'} key={'minutes'}>
                  {t('groups|minutes')}
                </MenuItem>
                <MenuItem value={'days'} key={'days'}>
                  {t('groups|days')}
                </MenuItem>
              </Field>
            </Box>
          </Grid>
        </Grid>
        <Grid item xs={12}>
          <Box mt={2} pl={2}>
            <FormLabel>{t('groups|Updates timeout')}</FormLabel>
            <Grid container spacing={2}>
              <Grid item xs={4}>
                <Field
                  name="updatesTimeout"
                  component={TextField}
                  margin="dense"
                  type="number"
                  defaultValue={values.updatesTimeout}
                  inputProps={{ min: 0 }}
                />
              </Grid>
              <Grid item xs={3}>
                <Box pr={2}>
                  <Field
                    name="updatesTimeoutUnit"
                    component={TextField}
                    margin="dense"
                    select
                    fullWidth
                    defaultValue={values.updatesTimeoutUnit}
                  >
                    <MenuItem value={'hours'} key={'hours'}>
                      {t('groups|hours')}
                    </MenuItem>
                    <MenuItem value={'minutes'} key={'minutes'}>
                      {t('groups|minutes')}
                    </MenuItem>
                    <MenuItem value={'days'} key={'days'}>
                      {t('groups|days')}
                    </MenuItem>
                  </Field>
                </Box>
              </Grid>
            </Grid>
          </Box>
        </Grid>
      </Box>
    </div>
  );
}
