import { render } from '@testing-library/react';
import React from 'react';
import Empty from '../../components/common/EmptyContent';

describe('Empty', () => {
  it('renders correct content', () => {
    const { asFragment } = render(<Empty>Some Dummy Content</Empty>);
    expect(asFragment()).toMatchSnapshot();
  });
});
