import { describe, expect, test } from 'vitest';

import { Channel } from '../../api/apiDataTypes';
import { theme } from '../../TestHelpers/theme';
import * as helpers from '../../utils/helpers';

describe('Utility Functions', () => {
  test('getKeyByValue should return correct key', () => {
    const obj = { a: 1, b: 2, c: 3 };
    expect(helpers.getKeyByValue(obj, 2)).toBe('b');
    expect(helpers.getKeyByValue(obj, 4)).toBeUndefined();
  });

  test('cleanSemverVersion should remove metadata', () => {
    expect(helpers.cleanSemverVersion('1.2.3+meta')).toBe('1.2.3');
    expect(helpers.cleanSemverVersion('1.2.3')).toBe('1.2.3');
  });

  test('getInstanceStatus returns correct status object', () => {
    const result = helpers.getInstanceStatus(3, '1.2.3');
    expect(result.status).toBe('Error');
    expect(result.explanation).toContain('1.2.3');
  });

  test('getErrorAndFlags extracts error message and flags', () => {
    const [errorMessages, flags] = helpers.getErrorAndFlags(3);
    expect(errorMessages).toContain('OmahaResponseHandlerError');
    expect(flags).toEqual([]);
  });

  describe('makeColorsForVersions', () => {
    test('assigns colors keyed by raw version strings without metadata', () => {
      const versions = ['2191.5.0', '2247.2.0'];
      const colors = helpers.makeColorsForVersions(theme, versions, null);
      expect(Object.keys(colors)).toContain('2191.5.0');
      expect(Object.keys(colors)).toContain('2247.2.0');
      expect(colors['2191.5.0']).toBeDefined();
      expect(colors['2247.2.0']).toBeDefined();
    });

    test('assigns colors keyed by raw version strings with build metadata', () => {
      const versions = ['3510.2.0+test', '2191.5.0'];
      const colors = helpers.makeColorsForVersions(theme, versions, null);
      expect(colors['3510.2.0+test']).toBeDefined();
      expect(colors['2191.5.0']).toBeDefined();
      expect(colors['3510.2.0']).toBeUndefined();
    });

    test('highlights channel package version with primary theme color for metadata versions', () => {
      const mockChannel = {
        package: {
          version: '3510.2.0+test',
        },
      } as Channel;
      const versions = ['3510.2.0+test', '2191.5.0'];
      const colors = helpers.makeColorsForVersions(theme, versions, mockChannel);
      expect(colors['3510.2.0+test']).toBe(theme.palette.primary.main);
      expect(colors['2191.5.0']).not.toBe(theme.palette.primary.main);
    });

    test('assigns distinct colors for multiple metadata variants of the same core version', () => {
      const versions = ['1.2.3+aws', '1.2.3+azure'];
      const colors = helpers.makeColorsForVersions(theme, versions, null);
      expect(colors['1.2.3+aws']).toBeDefined();
      expect(colors['1.2.3+azure']).toBeDefined();
      expect(colors['1.2.3+aws']).not.toBe(colors['1.2.3+azure']);
    });
  });
});
