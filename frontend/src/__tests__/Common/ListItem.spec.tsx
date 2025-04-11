import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import ListItem from '../../components/common/ListItem';

describe('List Item', () => {
  it('should render correct list item', () => {
    const { getByTestId } = render(<ListItem>{null}</ListItem>);
    expect(getByTestId('list-item')).toBeTruthy();
  });
});
