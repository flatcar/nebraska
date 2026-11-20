import { describe, expect, it } from 'vitest';

import { instanceModeOf } from '../../utils/instanceMode';

describe('instanceModeOf', () => {
  it.each([
    [{ instance_mode: 'control' }, 'control'],
    [{ instance_mode: 'edge' }, 'edge'],
    [{ instance_mode: 'single' }, 'single'],
    // An older server reports no mode, and every node was a single one.
    [{}, 'single'],
    [{ instance_mode: 'replica' }, 'single'],
  ])('reads %j as %s', (config, mode) => {
    expect(instanceModeOf(config)).toBe(mode);
  });
});
