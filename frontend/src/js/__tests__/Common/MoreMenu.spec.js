import { fireEvent, render } from '@testing-library/react';
import React from 'react';
import MoreMenu from '../../components/Common/MoreMenu';

describe('More Menu', () => {
  it('should render correct list of menu item', () => {
    const minProps = {
      options: [
        {
          label: 'item1',
        },
        {
          label: 'item2',
        },
      ],
    };
    const { getAllByTestId } = render(<MoreMenu options={minProps.options} />);
    expect(getAllByTestId('more-menu-item')).toHaveLength(2);
  });
});
