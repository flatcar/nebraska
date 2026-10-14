import HelpOutlineIcon from '@mui/icons-material/HelpOutline';
import {
  Box,
  Divider,
  FormControlLabel,
  FormHelperText,
  FormLabel,
  Grid,
  MenuItem,
  Radio,
  RadioGroup,
  Switch,
  Tooltip,
  Typography,
} from '@mui/material';
import { Field } from 'formik';
import { TextField } from 'formik-mui';
import { useTranslation } from 'react-i18next';

import TimezonePicker from '../../common/TimezonePicker';

export interface GroupPolicyFormProps {
  values: { [key: string]: string };
  setFieldValue: (formField: string, value: any) => any;
}

export default function GroupPolicyForm(props: GroupPolicyFormProps) {
  const { t } = useTranslation();
  const { values, setFieldValue } = props;

  return (
    <div style={{ padding: '1rem' }}>
      <Grid container justifyContent="space-between" spacing={4}>
        <Grid size={12}>
          <Box mt={1}>
            <FormLabel component="legend">{t('groups|update')}</FormLabel>
          </Box>
          <Grid container>
            <Grid size={6}>
              <FormControlLabel
                label={t('groups|updates_enabled')}
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
            <Grid size={6}>
              <FormControlLabel
                label={t('groups|safe_mode_lower')}
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
              <FormHelperText>{t('groups|update_policy_single_instance')}</FormHelperText>
            </Grid>
          </Grid>
        </Grid>
      </Grid>
      <Box mt={2}>
        <Divider />
      </Box>
      <Box mt={2}>
        <FormLabel component="legend">{t('groups|update_limits')}</FormLabel>
      </Box>
      <Box>
        <Grid container>
          <Grid size={6}>
            <FormControlLabel
              label={
                <Box display="flex" alignItems="center">
                  <Box pr={0.5}>{t('groups|office_hours_only_lower')}</Box>
                  <Box pt={0.1} color="#808080">
                    <Tooltip title={t('groups|update_policy_office_hours') || ''}>
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
          <Grid size={6}>
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
      <Box mt={2}>
        <FormLabel id="rollout-rate-label">{t('groups|rollout_rate')}</FormLabel>
        <RadioGroup
          aria-labelledby="rollout-rate-label"
          name="rolloutRate"
          value={values.updatesUnlimited ? 'unlimited' : 'limited'}
          onChange={(_e, value) => {
            setFieldValue('updatesUnlimited', value === 'unlimited');
          }}
        >
          <FormControlLabel
            value="unlimited"
            control={<Radio color="primary" />}
            label={t('groups|rollout_rate_unlimited')}
          />
          {values.updatesUnlimited && (
            <FormHelperText sx={{ mt: -1, ml: 4 }}>
              {values.safeMode
                ? t('groups|update_policy_no_rate_limit_safe_mode')
                : t('groups|update_policy_no_rate_limit')}
            </FormHelperText>
          )}
          <Box display="flex" alignItems="center" flexWrap="wrap" columnGap={1.5}>
            <FormControlLabel
              value="limited"
              control={<Radio color="primary" />}
              label={t('groups|rollout_rate_limited')}
            />
            {/* Formik keeps the hidden values, so the limit is restored when switching back. */}
            {!values.updatesUnlimited && (
              <>
                <Box width="7em">
                  <Field
                    name="maxUpdates"
                    component={TextField}
                    variant="standard"
                    margin="dense"
                    type="number"
                    fullWidth
                    inputProps={{ min: 1, 'aria-label': t('groups|max_updates') }}
                  />
                </Box>
                <Typography color="textSecondary">
                  {t('groups|updates_per', { count: Number(values.maxUpdates) })}
                </Typography>
                <Box width="5em">
                  <Field
                    name="updatesPeriodRange"
                    component={TextField}
                    variant="standard"
                    margin="dense"
                    type="number"
                    fullWidth
                    inputProps={{ min: 1, 'aria-label': t('groups|period_length') }}
                  />
                </Box>
                <Box width="7em">
                  <Field
                    name="updatesPeriodUnit"
                    component={TextField}
                    variant="standard"
                    margin="dense"
                    select
                    fullWidth
                    inputProps={{ 'aria-label': t('groups|period_unit') }}
                  >
                    <MenuItem value={'hours'} key={'hours'}>
                      {t('groups|time_hours')}
                    </MenuItem>
                    <MenuItem value={'minutes'} key={'minutes'}>
                      {t('groups|time_minutes')}
                    </MenuItem>
                    <MenuItem value={'days'} key={'days'}>
                      {t('groups|time_days')}
                    </MenuItem>
                  </Field>
                </Box>
              </>
            )}
          </Box>
        </RadioGroup>
      </Box>
      <Box my={2}>
        <Grid size={12}>
          <Box mt={2} pl={2}>
            <FormLabel>{t('groups|updates_timeout_lower')}</FormLabel>
            <Grid container spacing={2}>
              <Grid size={4}>
                <Field
                  name="updatesTimeout"
                  component={TextField}
                  variant="standard"
                  margin="dense"
                  type="number"
                  inputProps={{ min: 0 }}
                />
              </Grid>
              <Grid size={3}>
                <Box pr={2}>
                  <Field
                    name="updatesTimeoutUnit"
                    component={TextField}
                    variant="standard"
                    margin="dense"
                    select
                    fullWidth
                  >
                    <MenuItem value={'hours'} key={'hours'}>
                      {t('groups|time_hours')}
                    </MenuItem>
                    <MenuItem value={'minutes'} key={'minutes'}>
                      {t('groups|time_minutes')}
                    </MenuItem>
                    <MenuItem value={'days'} key={'days'}>
                      {t('groups|time_days')}
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
