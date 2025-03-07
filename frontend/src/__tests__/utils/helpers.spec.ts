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
});
