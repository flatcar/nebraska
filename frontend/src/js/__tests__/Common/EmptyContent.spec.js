import {render} from '@testing-library/react';
import React from 'react';
import Empty from '../../components/Common/EmptyContent';

describe('Empty', () => {
  it('renders correct content', () => {
    const {asFragment} = render(<Empty>Some Dummy Content</Empty>);
    expect(asFragment()).toMatchSnapshot();
  });
});
