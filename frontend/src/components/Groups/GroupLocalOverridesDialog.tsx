import {
  Box,
  Checkbox,
  Divider,
  FormControlLabel,
  FormHelperText,
  FormLabel,
  Switch,
  Typography,
} from '@mui/material';
import Button from '@mui/material/Button';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import { Field, Form, Formik, FormikErrors, FormikHelpers } from 'formik';
import { TextField } from 'formik-mui';
import { TFunction } from 'i18next';
import React from 'react';
import { useTranslation } from 'react-i18next';

import { Group, GroupLocalOverrides } from '../../api/apiDataTypes';
import { applicationsStore } from '../../stores/Stores';
import { effectivePolicy } from '../../utils/groupPolicy';
import { MAX_UPDATES_PER_TIME_PERIOD } from '../../utils/helpers';
import TimezonePicker, { DEFAULT_TIMEZONE } from '../common/TimezonePicker';
import { formatUpdateLimits } from './GroupItem';
import RolloutRateFields, { TIME_UNITS, UnitField } from './RolloutRateFields';

export interface LocalOverridesValues {
  safeModeOverride: boolean;
  safeMode: boolean;
  officeHoursOverride: boolean;
  onlyOfficeHours: boolean;
  timezone: string;
  rateOverride: boolean;
  updatesUnlimited: boolean;
  maxUpdates: number | string;
  updatesPeriodRange: number | string;
  updatesPeriodUnit: string;
  timeoutOverride: boolean;
  updatesTimeout: number | string;
  updatesTimeoutUnit: string;
}

// The most minutes, hours or days the server takes in an interval override.
export const MAX_INTERVAL_COUNT = 999999;

// Reads an interval written as "<n> <unit>", the form the group dialogs write.
// The count is read as a number, so one written with leading zeros is sent in
// the form the server takes.
function splitInterval(interval: string): [string, string] {
  const match = /^\s*(\d+)\s+(minute|hour|day)s?\s*$/i.exec(interval);
  if (!match) {
    return ['', ''];
  }
  return [String(Number(match[1])), match[2].toLowerCase() + 's'];
}

export function initialLocalOverridesValues(group: Group): LocalOverridesValues {
  const overrides = group.local_overrides;
  const has = (field: keyof GroupLocalOverrides) =>
    overrides?.[field] !== null && overrides?.[field] !== undefined;
  // A section that is not overridden starts from the default, which is then
  // also what applies here.
  const effective = effectivePolicy(group);
  const unlimited = effective.policy_max_updates_per_period >= MAX_UPDATES_PER_TIME_PERIOD;
  const [periodRange, periodUnit] = splitInterval(effective.policy_period_interval);
  const [timeout, timeoutUnit] = splitInterval(effective.policy_update_timeout);

  return {
    safeModeOverride: has('policy_safe_mode'),
    safeMode: effective.policy_safe_mode,
    officeHoursOverride: has('policy_office_hours') || has('policy_timezone'),
    onlyOfficeHours: effective.policy_office_hours,
    timezone: effective.policy_timezone || DEFAULT_TIMEZONE,
    rateOverride: has('policy_max_updates_per_period') || has('policy_period_interval'),
    updatesUnlimited: unlimited,
    // An unlimited rate has no limit to show, so start a limit at 1 per hour.
    maxUpdates: unlimited ? 1 : effective.policy_max_updates_per_period,
    updatesPeriodRange: unlimited ? 1 : periodRange,
    updatesPeriodUnit: unlimited ? 'hours' : periodUnit,
    timeoutOverride: has('policy_update_timeout'),
    updatesTimeout: timeout,
    updatesTimeoutUnit: timeoutUnit,
  };
}

function interval(count: number | string, unit: string) {
  return `${Number(String(count).trim())} ${unit}`;
}

// The patch carries only what the user changed, because the server keeps every
// override a patch leaves out: saving one setting cannot undo a pause the node
// set while the dialog was open. A section the user starts to override sends
// every setting in it, as shown. A section that was already overridden sends
// only the settings that changed, so one it inherits keeps following the
// default.
export function localOverridesPatch(
  initial: LocalOverridesValues,
  values: LocalOverridesValues
): Partial<GroupLocalOverrides> {
  const changed = (...keys: (keyof LocalOverridesValues)[]) =>
    keys.some(key => String(values[key]) !== String(initial[key]));
  const patch: Partial<GroupLocalOverrides> = {};

  if (changed('safeModeOverride') || (values.safeModeOverride && changed('safeMode'))) {
    patch.policy_safe_mode = values.safeModeOverride ? values.safeMode : null;
  }

  if (!values.officeHoursOverride) {
    if (initial.officeHoursOverride) {
      patch.policy_office_hours = null;
      patch.policy_timezone = null;
    }
  } else if (changed('officeHoursOverride', 'onlyOfficeHours')) {
    patch.policy_office_hours = values.onlyOfficeHours;
    // The timezone only matters while office hours apply.
    patch.policy_timezone = values.onlyOfficeHours ? values.timezone : null;
  } else if (values.onlyOfficeHours && changed('timezone')) {
    patch.policy_timezone = values.timezone;
  }

  if (!values.rateOverride) {
    if (initial.rateOverride) {
      patch.policy_max_updates_per_period = null;
      patch.policy_period_interval = null;
    }
  } else if (values.updatesUnlimited) {
    if (changed('rateOverride', 'updatesUnlimited')) {
      patch.policy_max_updates_per_period = MAX_UPDATES_PER_TIME_PERIOD;
      // Safe mode still counts updates per period, so keep the default's.
      patch.policy_period_interval = null;
    }
  } else {
    const whole = changed('rateOverride', 'updatesUnlimited');
    if (whole || changed('maxUpdates')) {
      patch.policy_max_updates_per_period = Number(String(values.maxUpdates).trim());
    }
    if (whole || changed('updatesPeriodRange', 'updatesPeriodUnit')) {
      patch.policy_period_interval = interval(values.updatesPeriodRange, values.updatesPeriodUnit);
    }
  }

  if (
    changed('timeoutOverride') ||
    (values.timeoutOverride && changed('updatesTimeout', 'updatesTimeoutUnit'))
  ) {
    patch.policy_update_timeout = values.timeoutOverride
      ? interval(values.updatesTimeout, values.updatesTimeoutUnit)
      : null;
  }

  return patch;
}

function countError(t: TFunction, value: number | string, max: number, tooHigh: string) {
  const text = String(value).trim();
  if (text === '') {
    return t('frequent|required');
  }
  if (!/^-?\d+$/.test(text)) {
    return t('common|integer_number_error');
  }
  const count = Number(text);
  if (count < 1) {
    return t('common|min_value_error', { number: 1 });
  }
  return count > max ? tooHigh : undefined;
}

// Checks only what the patch sends, so a section the user did not touch cannot
// block saving another, even when the form cannot show its value.
export function validateLocalOverrides(
  t: TFunction,
  initial: LocalOverridesValues,
  values: LocalOverridesValues
): FormikErrors<LocalOverridesValues> {
  const patch = localOverridesPatch(initial, values);
  const errors: FormikErrors<LocalOverridesValues> = {};
  const intervalTooHigh = t('groups|interval_too_high', { number: MAX_INTERVAL_COUNT });

  function check(
    key: 'maxUpdates' | 'updatesPeriodRange' | 'updatesTimeout',
    max: number,
    tooHigh: string
  ) {
    const error = countError(t, values[key], max, tooHigh);
    if (error) {
      errors[key] = error;
    }
  }
  function checkUnit(key: 'updatesPeriodUnit' | 'updatesTimeoutUnit') {
    if (!TIME_UNITS.includes(values[key])) {
      errors[key] = t('frequent|required');
    }
  }

  if (typeof patch.policy_max_updates_per_period === 'number' && !values.updatesUnlimited) {
    check(
      'maxUpdates',
      MAX_UPDATES_PER_TIME_PERIOD - 1,
      t('groups|max_updates_too_high', { number: MAX_UPDATES_PER_TIME_PERIOD })
    );
  }
  if (typeof patch.policy_period_interval === 'string') {
    check('updatesPeriodRange', MAX_INTERVAL_COUNT, intervalTooHigh);
    checkUnit('updatesPeriodUnit');
  }
  if (typeof patch.policy_update_timeout === 'string') {
    check('updatesTimeout', MAX_INTERVAL_COUNT, intervalTooHigh);
    checkUnit('updatesTimeoutUnit');
  }

  return errors;
}

async function errorMessage(error: unknown): Promise<string | null> {
  if (error instanceof Response && error.status === 400) {
    try {
      const body = await error.json();
      return body.message || null;
    } catch {
      return null;
    }
  }
  return null;
}

interface SectionProps {
  title: string;
  defaultValue: string;
  override: {
    checked: boolean;
    onChange: (checked: boolean) => void;
  };
  children: React.ReactNode;
}

function Section({ title, defaultValue, override, children }: SectionProps) {
  const { t } = useTranslation();

  return (
    <Box py={1.5} role="group" aria-label={title}>
      <Box display="flex" alignItems="center" justifyContent="space-between" gap={2}>
        <Box>
          <FormLabel component="legend">{title}</FormLabel>
          <Typography color="textSecondary" sx={{ fontSize: '0.875rem' }}>
            {t('groups|default_value', { value: defaultValue })}
          </Typography>
        </Box>
        <FormControlLabel
          label={t('groups|override')}
          control={
            <Checkbox
              checked={override.checked}
              onChange={e => override.onChange(e.target.checked)}
              inputProps={{ 'aria-label': t('groups|override_setting', { setting: title }) }}
            />
          }
        />
      </Box>
      {override.checked && <Box mt={1}>{children}</Box>}
    </Box>
  );
}

export interface GroupLocalOverridesDialogProps {
  group: Group;
  show: boolean;
  onHide: () => void;
}

// Edits this node's overrides of the group default. Overrides stay on this node
// and are not replicated.
export default function GroupLocalOverridesDialog(props: GroupLocalOverridesDialogProps) {
  const { show, onHide } = props;
  const { t } = useTranslation();
  // Each opening gets a new form, which starts from the group as it is then,
  // even when the dialog reopens before it has finished closing.
  const [openings, setOpenings] = React.useState(0);
  const [wasShown, setWasShown] = React.useState(show);
  if (show !== wasShown) {
    setWasShown(show);
    if (show) {
      setOpenings(openings + 1);
    }
  }

  return (
    <Dialog open={show} onClose={onHide} aria-labelledby="local-overrides-dialog-title">
      <DialogTitle id="local-overrides-dialog-title">
        {t('groups|local_overrides_title')}
      </DialogTitle>
      <LocalOverridesForm key={openings} group={props.group} onHide={onHide} />
    </Dialog>
  );
}

function LocalOverridesForm(props: { group: Group; onHide: () => void }) {
  const { group, onHide } = props;
  const { t } = useTranslation();
  // What the user saw when the dialog opened, which the patch is taken
  // against, so a change made elsewhere while the dialog is open, such as by
  // the control node, is not sent back.
  const [initialValues] = React.useState(() => initialLocalOverridesValues(group));

  // The default is what a section falls back to, so it reads from the group
  // fields, which on a distributed node are the default.
  const yesNo = (value: boolean) => (value ? t('frequent|yes') : t('frequent|no'));
  const defaultOfficeHours = group.policy_office_hours
    ? t('groups|office_hours_in', { timezone: group.policy_timezone })
    : t('frequent|no');

  async function handleSubmit(
    values: LocalOverridesValues,
    actions: FormikHelpers<LocalOverridesValues>
  ) {
    const patch = localOverridesPatch(initialValues, values);
    if (Object.keys(patch).length === 0) {
      onHide();
      return;
    }

    try {
      await applicationsStore().updateGroupLocalOverrides(group.application_id, group.id, patch);
      onHide();
    } catch (error) {
      actions.setStatus({
        statusMessage: (await errorMessage(error)) || t('common|generic_error'),
      });
    }
  }

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={handleSubmit}
      validate={values => validateLocalOverrides(t, initialValues, values)}
    >
      {({ values, status, setFieldValue, isSubmitting }) => {
        // Turning an override off shows the default again, which is also
        // where turning it back on starts from.
        function toggle(flag: keyof LocalOverridesValues, fields: (keyof LocalOverridesValues)[]) {
          return (checked: boolean) => {
            setFieldValue(flag, checked);
            if (!checked) {
              const defaults = initialLocalOverridesValues({ ...group, local_overrides: null });
              fields.forEach(field => setFieldValue(field, defaults[field]));
            }
          };
        }

        return (
          <Form data-testid="local-overrides-form" noValidate>
            <DialogContent sx={{ pt: 0 }}>
              <DialogContentText sx={{ fontSize: '0.875rem' }}>
                {t('groups|local_overrides_intro')}
              </DialogContentText>
              {status && status.statusMessage && (
                <DialogContentText color="error" role="alert" sx={{ mt: 1 }}>
                  {status.statusMessage}
                </DialogContentText>
              )}
              <Section
                title={t('groups|safe_mode')}
                defaultValue={yesNo(group.policy_safe_mode)}
                override={{
                  checked: values.safeModeOverride,
                  onChange: toggle('safeModeOverride', ['safeMode']),
                }}
              >
                <FormControlLabel
                  label={t('groups|safe_mode_lower')}
                  control={
                    <Switch
                      checked={values.safeMode}
                      color="primary"
                      onChange={e => setFieldValue('safeMode', e.target.checked)}
                    />
                  }
                />
                <FormHelperText>{t('groups|update_policy_single_instance')}</FormHelperText>
              </Section>
              <Divider />
              <Section
                title={t('groups|office_hours_only')}
                defaultValue={defaultOfficeHours}
                override={{
                  checked: values.officeHoursOverride,
                  onChange: toggle('officeHoursOverride', ['onlyOfficeHours', 'timezone']),
                }}
              >
                <FormControlLabel
                  label={t('groups|office_hours_only_lower')}
                  control={
                    <Switch
                      checked={values.onlyOfficeHours}
                      color="primary"
                      onChange={e => setFieldValue('onlyOfficeHours', e.target.checked)}
                    />
                  }
                />
                {values.onlyOfficeHours && (
                  <Box maxWidth="20em" mt={1}>
                    <TimezonePicker
                      value={values.timezone}
                      onSelect={timezone => setFieldValue('timezone', timezone)}
                    />
                  </Box>
                )}
              </Section>
              <Divider />
              <Section
                title={t('groups|rollout_rate')}
                defaultValue={formatUpdateLimits(t, group)}
                override={{
                  checked: values.rateOverride,
                  onChange: toggle('rateOverride', [
                    'updatesUnlimited',
                    'maxUpdates',
                    'updatesPeriodRange',
                    'updatesPeriodUnit',
                  ]),
                }}
              >
                <RolloutRateFields />
              </Section>
              <Divider />
              <Section
                title={t('groups|updates_timeout')}
                defaultValue={group.policy_update_timeout}
                override={{
                  checked: values.timeoutOverride,
                  onChange: toggle('timeoutOverride', ['updatesTimeout', 'updatesTimeoutUnit']),
                }}
              >
                <Box display="flex" columnGap={2}>
                  <Box width="7em">
                    <Field
                      name="updatesTimeout"
                      component={TextField}
                      variant="standard"
                      margin="dense"
                      type="number"
                      fullWidth
                      inputProps={{ min: 1, 'aria-label': t('groups|updates_timeout_lower') }}
                    />
                  </Box>
                  <Box width="7em">
                    <UnitField name="updatesTimeoutUnit" label={t('groups|timeout_unit')} />
                  </Box>
                </Box>
              </Section>
            </DialogContent>
            <DialogActions>
              <Button onClick={onHide}>{t('frequent|cancel')}</Button>
              <Button type="submit" variant="contained" disabled={isSubmitting}>
                {t('frequent|save')}
              </Button>
            </DialogActions>
          </Form>
        );
      }}
    </Formik>
  );
}
