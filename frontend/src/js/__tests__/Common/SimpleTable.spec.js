import { render } from '@testing-library/react';
import React from 'react';
import SimpleTable from '../../components/Common/SimpleTable';

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
    expect(getByText(minProps.instances[0].key1)).toBeInTheDocument();
    expect(getByText(minProps.instances[0].key2)).toBeInTheDocument();
    expect(getByText(minProps.columns.key1)).toBeInTheDocument();
    expect(getByText(minProps.columns.key2)).toBeInTheDocument();
  });
});
