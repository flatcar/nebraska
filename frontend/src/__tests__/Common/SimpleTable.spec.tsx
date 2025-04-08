import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import SimpleTable from '../../components/common/SimpleTable';

describe('Simple Table', () => {
  const minProps = {
    instances: [
      {
        key1: 'value1',
        key2: 'value2',
      },
    ],
    columns: {
      key1: 'Column1',
      key2: 'Column2',
    },
  };
  it('shoudl render Table with correct data', async () => {
    const { getByText } = render(<SimpleTable {...minProps} />);
    expect(getByText(minProps.instances[0].key1)).toBeTruthy();
    expect(getByText(minProps.instances[0].key2)).toBeTruthy();
    expect(getByText(minProps.columns.key1)).toBeTruthy();
    expect(getByText(minProps.columns.key2)).toBeTruthy();
  });
});
