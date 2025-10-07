import { describe, expect, test } from 'vitest';

import { REGEX_REVERSE_DOMAIN_ID } from '../../utils/regex';

describe('reverseDomainIdRegex', () => {
  const regex = REGEX_REVERSE_DOMAIN_ID;

  test.each([
    // valid examples
    ['com.example', true],
    ['org.my-app', true],
    ['net.service1', true],
    ['a.b', true],
    ['A.B1', true],
    ['abc.def-ghi', true],
    ['MyCompany.MyApp', true],

    // invalid examples
    ['a', false], // only one segment
    ['com.', false], // ends with a dot
    ['.example', false], // starts with a dot
    ['com..example', false], // double dots
    ['com.-example', false], // segment starts with dash
    ['com.exa-', false], // segment ends with dash
    ['1com.example', false], // first segment starts with digit
    ['com.exa$mple', false], // invalid character
    ['com.exa mple', false], // space not allowed
    ['com..example', false], // empty segment
  ])('validates "%s"', (input, expected) => {
    expect(regex.test(input)).toBe(expected);
  });
});
