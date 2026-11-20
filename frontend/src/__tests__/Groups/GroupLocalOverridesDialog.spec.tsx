import { TFunction } from 'i18next';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mockUpdateGroupLocalOverrides = vi.fn();

vi.mock('../../stores/Stores', () => ({
  applicationsStore: vi.fn(() => ({
    updateGroupLocalOverrides: mockUpdateGroupLocalOverrides,
  })),
}));

import '../../i18n/config.ts';

import { StyledEngineProvider, ThemeProvider } from '@mui/material/styles';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';

import { Group } from '../../api/apiDataTypes';
import GroupLocalOverridesDialog, {
  initialLocalOverridesValues,
  localOverridesPatch,
  validateLocalOverrides,
} from '../../components/Groups/GroupLocalOverridesDialog';
import { makeGroup } from '../../TestHelpers/groups';
import { theme } from '../../TestHelpers/theme';
import { MAX_UPDATES_PER_TIME_PERIOD } from '../../utils/helpers';

const onHide = vi.fn();

function dialog(group: Group, show = true) {
  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider theme={theme}>
        <GroupLocalOverridesDialog group={group} show={show} onHide={onHide} />
      </ThemeProvider>
    </StyledEngineProvider>
  );
}

function spinbutton(name: string) {
  return screen.getByRole('spinbutton', { name }) as HTMLInputElement;
}

async function type(name: string, value: string) {
  await act(async () => fireEvent.change(spinbutton(name), { target: { value } }));
}

async function click(label: string) {
  await act(async () => fireEvent.click(screen.getByLabelText(label)));
}

async function save() {
  await act(async () => fireEvent.click(screen.getByRole('button', { name: 'Save' })));
}

function sentPatch() {
  expect(mockUpdateGroupLocalOverrides).toHaveBeenCalledTimes(1);
  const [appID, groupID, patch] = mockUpdateGroupLocalOverrides.mock.calls[0];
  expect([appID, groupID]).toEqual(['app-1', 'group-1']);
  return patch;
}

describe('GroupLocalOverridesDialog', () => {
  beforeEach(() => {
    mockUpdateGroupLocalOverrides.mockReset().mockResolvedValue(undefined);
    onHide.mockReset();
  });

  it('shows the default each setting falls back to', () => {
    render(dialog(makeGroup({ policy_updates_enabled: false })));

    expect(screen.getAllByText('Default: No')).toHaveLength(2);
    expect(screen.getByText('Default: Max 2 / 15 minutes')).toBeTruthy();
    expect(screen.getByText('Default: 60 minutes')).toBeTruthy();
    // Controls appear once a setting is overridden.
    expect(screen.queryByLabelText('Safe mode')).toBeNull();
    // Pausing has its own button on the page.
    expect(screen.queryByText(/Pause/)).toBeNull();
  });

  it('tells that an override can loosen the default', () => {
    render(dialog(makeGroup()));

    expect(screen.getByText(/stricter or looser than the default/)).toBeTruthy();
  });

  it('closes without a request when nothing changed', async () => {
    render(dialog(makeGroup({ policy_safe_mode: true })));
    await save();

    await waitFor(() => expect(onHide).toHaveBeenCalled());
    expect(mockUpdateGroupLocalOverrides).not.toHaveBeenCalled();
  });

  it('sends only the setting that changed', async () => {
    render(dialog(makeGroup({ policy_updates_enabled: false, policy_update_timeout: '2 hours' })));
    await click('Override Safe Mode');
    await click('Safe mode');
    await save();

    await waitFor(() => expect(onHide).toHaveBeenCalled());
    expect(sentPatch()).toEqual({ policy_safe_mode: true });
  });

  it('removes an override that is turned off', async () => {
    render(
      dialog(makeGroup({ policy_max_updates_per_period: 5, policy_period_interval: '2 hours' }))
    );
    expect(spinbutton('Max number of updates').value).toBe('5');

    await click('Override Rollout rate');
    expect(screen.queryByRole('spinbutton', { name: 'Max number of updates' })).toBeNull();
    await save();

    await waitFor(() => expect(onHide).toHaveBeenCalled());
    expect(sentPatch()).toEqual({
      policy_max_updates_per_period: null,
      policy_period_interval: null,
    });
  });

  it('keeps an override set elsewhere while the dialog is open', async () => {
    const { rerender } = render(dialog(makeGroup()));
    // The store refreshes the group, which meanwhile got a timeout override.
    rerender(dialog(makeGroup({ policy_update_timeout: '2 hours' })));

    await click('Override Safe Mode');
    await save();

    await waitFor(() => expect(onHide).toHaveBeenCalled());
    expect(sentPatch()).toEqual({ policy_safe_mode: false });
  });

  it('starts from the group as it is each time it opens', async () => {
    const { rerender } = render(dialog(makeGroup()));
    expect(screen.queryByRole('spinbutton', { name: 'Updates timeout' })).toBeNull();

    // Reopened before it has finished closing, with a timeout override set meanwhile.
    rerender(dialog(makeGroup(), false));
    rerender(dialog(makeGroup({ policy_update_timeout: '2 hours' })));

    expect(spinbutton('Updates timeout').value).toBe('2');
  });

  it('saves a setting while another section holds a value it cannot show', async () => {
    // The node inherits a period the form cannot show.
    render(
      dialog({ ...makeGroup({ policy_max_updates_per_period: 5 }), policy_period_interval: 'PT2H' })
    );
    expect(spinbutton('Period length').value).toBe('');

    await click('Override Safe Mode');
    await save();

    await waitFor(() => expect(onHide).toHaveBeenCalled());
    expect(sentPatch()).toEqual({ policy_safe_mode: false });
  });

  it('refuses an invalid rate before sending it', async () => {
    render(dialog(makeGroup()));
    await click('Override Rollout rate');

    // The backend reads this many as unlimited.
    await type('Max number of updates', '900000');
    await save();
    expect(await screen.findByText(/Must be less than 900,000/)).toBeTruthy();

    await type('Max number of updates', '0');
    await save();
    expect(await screen.findByText('Must be greater than or equal to 1')).toBeTruthy();

    await type('Max number of updates', '');
    await save();
    expect(await screen.findByText('Required')).toBeTruthy();

    await type('Max number of updates', '3');
    await type('Period length', '1000000');
    await save();
    expect(await screen.findByText('Must be at most 999,999')).toBeTruthy();

    expect(mockUpdateGroupLocalOverrides).not.toHaveBeenCalled();
  });

  it('shows why the node refused the change', async () => {
    mockUpdateGroupLocalOverrides.mockRejectedValue(
      new Response(
        JSON.stringify({
          message: 'a policy_timezone override needs a policy_office_hours override',
        }),
        { status: 400 }
      )
    );
    render(dialog(makeGroup()));
    await click('Override Safe Mode');
    await save();

    expect((await screen.findByRole('alert')).textContent).toBe(
      'a policy_timezone override needs a policy_office_hours override'
    );
    expect(onHide).not.toHaveBeenCalled();
  });
});

// Returns the key of each message, which tells which check refused a value.
const t = ((key: string) => key) as unknown as TFunction;

describe('localOverridesPatch', () => {
  it('sends nothing when nothing changed', () => {
    const initial = initialLocalOverridesValues(
      makeGroup({ policy_updates_enabled: false, policy_safe_mode: true })
    );
    expect(localOverridesPatch(initial, { ...initial })).toEqual({});
  });

  it('sends only the sections that changed', () => {
    const initial = initialLocalOverridesValues(makeGroup());

    expect(
      localOverridesPatch(initial, { ...initial, safeModeOverride: true, safeMode: true })
    ).toEqual({ policy_safe_mode: true });
  });

  it('removes an override that is turned off', () => {
    const initial = initialLocalOverridesValues(
      makeGroup({ policy_max_updates_per_period: 5, policy_period_interval: '2 hours' })
    );
    expect(initial.rateOverride).toBe(true);
    expect(initial.maxUpdates).toBe(5);

    expect(localOverridesPatch(initial, { ...initial, rateOverride: false })).toEqual({
      policy_max_updates_per_period: null,
      policy_period_interval: null,
    });
  });

  it('pins every setting of a section it starts to override', () => {
    const initial = initialLocalOverridesValues(makeGroup());

    expect(
      localOverridesPatch(initial, {
        ...initial,
        rateOverride: true,
        maxUpdates: '7',
        updatesPeriodRange: '3',
        updatesPeriodUnit: 'hours',
      })
    ).toEqual({ policy_max_updates_per_period: 7, policy_period_interval: '3 hours' });

    // As shown, even where the user kept the default's value.
    expect(localOverridesPatch(initial, { ...initial, rateOverride: true })).toEqual({
      policy_max_updates_per_period: 2,
      policy_period_interval: '15 minutes',
    });

    expect(
      localOverridesPatch(initial, { ...initial, rateOverride: true, updatesUnlimited: true })
    ).toEqual({
      policy_max_updates_per_period: MAX_UPDATES_PER_TIME_PERIOD,
      policy_period_interval: null,
    });
  });

  it('leaves a setting a section inherits to the default', () => {
    // The node overrides the number of updates only, and follows the default's period.
    const initial = initialLocalOverridesValues(makeGroup({ policy_max_updates_per_period: 5 }));
    expect(initial.rateOverride).toBe(true);
    expect([initial.updatesPeriodRange, initial.updatesPeriodUnit]).toEqual(['15', 'minutes']);

    expect(localOverridesPatch(initial, { ...initial, maxUpdates: '9' })).toEqual({
      policy_max_updates_per_period: 9,
    });
    expect(localOverridesPatch(initial, { ...initial, updatesPeriodRange: '30' })).toEqual({
      policy_period_interval: '30 minutes',
    });
  });

  it('switches an overridden rate between limited and unlimited', () => {
    const limited = initialLocalOverridesValues(
      makeGroup({ policy_max_updates_per_period: 5, policy_period_interval: '2 hours' })
    );
    expect(localOverridesPatch(limited, { ...limited, updatesUnlimited: true })).toEqual({
      policy_max_updates_per_period: MAX_UPDATES_PER_TIME_PERIOD,
      policy_period_interval: null,
    });

    const unlimited = initialLocalOverridesValues(
      makeGroup({ policy_max_updates_per_period: MAX_UPDATES_PER_TIME_PERIOD })
    );
    expect(unlimited.updatesUnlimited).toBe(true);
    expect(localOverridesPatch(unlimited, { ...unlimited, updatesUnlimited: false })).toEqual({
      policy_max_updates_per_period: 1,
      policy_period_interval: '1 hours',
    });
  });

  it('writes office hours with their timezone', () => {
    const initial = initialLocalOverridesValues(makeGroup());

    expect(
      localOverridesPatch(initial, {
        ...initial,
        officeHoursOverride: true,
        onlyOfficeHours: true,
        timezone: 'Asia/Tokyo',
      })
    ).toEqual({ policy_office_hours: true, policy_timezone: 'Asia/Tokyo' });

    expect(localOverridesPatch(initial, { ...initial, officeHoursOverride: true })).toEqual({
      policy_office_hours: false,
      policy_timezone: null,
    });
  });

  it('changes office hours that were already overridden', () => {
    const initial = initialLocalOverridesValues(
      makeGroup({ policy_office_hours: true, policy_timezone: 'Asia/Tokyo' })
    );

    expect(localOverridesPatch(initial, { ...initial, timezone: 'Europe/Paris' })).toEqual({
      policy_timezone: 'Europe/Paris',
    });
    expect(localOverridesPatch(initial, { ...initial, onlyOfficeHours: false })).toEqual({
      policy_office_hours: false,
      policy_timezone: null,
    });
    expect(localOverridesPatch(initial, { ...initial, officeHoursOverride: false })).toEqual({
      policy_office_hours: null,
      policy_timezone: null,
    });
  });

  it('writes the timeout with its unit', () => {
    const initial = initialLocalOverridesValues(makeGroup());
    expect(initial.updatesTimeout).toBe('60');
    expect(initial.updatesTimeoutUnit).toBe('minutes');

    expect(
      localOverridesPatch(initial, { ...initial, timeoutOverride: true, updatesTimeout: 2 })
    ).toEqual({ policy_update_timeout: '2 minutes' });
  });

  it('sends a count written with leading zeros in the form the server takes', () => {
    const group = { ...makeGroup(), policy_update_timeout: '060 minutes' };
    const initial = initialLocalOverridesValues(group);
    expect(initial.updatesTimeout).toBe('60');

    expect(localOverridesPatch(initial, { ...initial, timeoutOverride: true })).toEqual({
      policy_update_timeout: '60 minutes',
    });
    expect(
      localOverridesPatch(initial, { ...initial, timeoutOverride: true, updatesTimeout: '007' })
    ).toEqual({ policy_update_timeout: '7 minutes' });
  });

  it('reads singular units and leaves other intervals blank', () => {
    const group = makeGroup({ policy_update_timeout: '1 hour', policy_period_interval: 'PT2H' });
    const initial = initialLocalOverridesValues(group);

    expect([initial.updatesTimeout, initial.updatesTimeoutUnit]).toEqual(['1', 'hours']);
    expect([initial.updatesPeriodRange, initial.updatesPeriodUnit]).toEqual(['', '']);
  });
});

describe('validateLocalOverrides', () => {
  it('checks only what the patch sends', () => {
    // The node overrides the number of updates and inherits a default period
    // the form cannot show, which stays out of the patch and the check.
    const group = {
      ...makeGroup({ policy_max_updates_per_period: 5 }),
      policy_period_interval: 'PT2H',
    };
    const initial = initialLocalOverridesValues(group);
    expect([initial.updatesPeriodRange, initial.updatesPeriodUnit]).toEqual(['', '']);

    expect(
      validateLocalOverrides(t, initial, { ...initial, safeModeOverride: true, safeMode: true })
    ).toEqual({});
    expect(validateLocalOverrides(t, initial, { ...initial, maxUpdates: '9' })).toEqual({});

    expect(validateLocalOverrides(t, initial, { ...initial, updatesPeriodUnit: 'hours' })).toEqual({
      updatesPeriodRange: 'frequent|required',
    });
  });

  it('asks for a period when a rate starts from one the form cannot show', () => {
    const initial = initialLocalOverridesValues({
      ...makeGroup(),
      policy_period_interval: 'PT2H',
    });

    expect(validateLocalOverrides(t, initial, { ...initial, rateOverride: true })).toEqual({
      updatesPeriodRange: 'frequent|required',
      updatesPeriodUnit: 'frequent|required',
    });
  });

  it.each([
    ['', 'frequent|required'],
    ['1.5', 'common|integer_number_error'],
    ['abc', 'common|integer_number_error'],
    ['0', 'common|min_value_error'],
    ['-2', 'common|min_value_error'],
    ['1000000', 'groups|interval_too_high'],
    ['999999', undefined],
    [' 3 ', undefined],
  ])('takes %j minutes as a timeout: %s', (count, error) => {
    const initial = initialLocalOverridesValues(makeGroup());
    const values = { ...initial, timeoutOverride: true, updatesTimeout: count };

    expect(validateLocalOverrides(t, initial, values).updatesTimeout).toBe(error);
  });

  it('refuses as many updates as the server reads as unlimited', () => {
    const initial = initialLocalOverridesValues(makeGroup());
    const limit = (maxUpdates: string) =>
      validateLocalOverrides(t, initial, { ...initial, rateOverride: true, maxUpdates }).maxUpdates;

    expect(limit(String(MAX_UPDATES_PER_TIME_PERIOD))).toBe('groups|max_updates_too_high');
    expect(limit(String(MAX_UPDATES_PER_TIME_PERIOD - 1))).toBeUndefined();
  });

  it('does not check a number for an unlimited rate', () => {
    const initial = initialLocalOverridesValues(makeGroup());

    expect(
      validateLocalOverrides(t, initial, {
        ...initial,
        rateOverride: true,
        updatesUnlimited: true,
        maxUpdates: '',
        updatesPeriodRange: '',
      })
    ).toEqual({});
  });
});
