import '../../i18n/config.ts';

import { render } from '@testing-library/react';
import React from 'react';
import { describe, expect, it } from 'vitest';

import AutoCompletePicker from '../../components/common/AutoCompletePicker';

describe('AutoCompletePicker', () => {
  it('does not keep calling onBottomScrolled when the parent re-renders', () => {
    let calls = 0;

    // Like ChannelEdit: loading more re-renders the parent with a new callback.
    function Parent() {
      const [page, setPage] = React.useState(0);
      return (
        <AutoCompletePicker
          initialOpen
          defaultValue=""
          label="Package"
          placeholder="Pick a package"
          dialogTitle="Choose a package"
          pickerPlaceholder="Start typing"
          suggestions={[{ primary: '1.0.0', secondary: '' }]}
          onSelect={() => {}}
          onValueChanged={() => {}}
          onBottomScrolled={() => {
            calls++;
            if (page < 100) setPage(page + 1);
          }}
        />
      );
    }

    render(<Parent />);
    expect(calls).toBe(1);
  });
});
