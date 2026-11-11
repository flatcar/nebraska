import {
  Box,
  FormControlLabel,
  FormHelperText,
  MenuItem,
  Radio,
  RadioGroup,
  Typography,
} from '@mui/material';
import { Field, useFormikContext } from 'formik';
import { TextField } from 'formik-mui';
import { useTranslation } from 'react-i18next';

export const TIME_UNITS = ['hours', 'minutes', 'days'];

interface RolloutRateValues {
  safeMode: boolean;
  updatesUnlimited: boolean;
  maxUpdates: number | string;
}

export function UnitField(props: { name: string; label: string }) {
  const { t } = useTranslation();
  const labels: Record<string, string> = {
    hours: t('groups|time_hours'),
    minutes: t('groups|time_minutes'),
    days: t('groups|time_days'),
  };

  return (
    <Field
      name={props.name}
      component={TextField}
      variant="standard"
      margin="dense"
      select
      fullWidth
      inputProps={{ 'aria-label': props.label }}
    >
      {TIME_UNITS.map(unit => (
        <MenuItem value={unit} key={unit}>
          {labels[unit]}
        </MenuItem>
      ))}
    </Field>
  );
}

export default function RolloutRateFields() {
  const { t } = useTranslation();
  const { values, setFieldValue } = useFormikContext<RolloutRateValues>();

  return (
    <RadioGroup
      aria-label={t('groups|rollout_rate')}
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
              <UnitField name="updatesPeriodUnit" label={t('groups|period_unit')} />
            </Box>
          </>
        )}
      </Box>
    </RadioGroup>
  );
}
