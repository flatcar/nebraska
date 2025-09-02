import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import MoreMenu from '../../components/common/MoreMenu';

describe('More Menu', () => {
  it('should render correct list of menu item', () => {
    const minProps = {
      options: [
        {
          label: 'item1',
          action: () => {},
        },
        {
          label: 'item2',
          action: () => {},
        },
      ],
    };
    const { getAllByTestId } = render(<MoreMenu options={minProps.options} />);
    expect(getAllByTestId('more-menu-item')).toHaveLength(2);
  });
});
