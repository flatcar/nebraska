import {render} from '@testing-library/react';
import React from 'react';
import Empty from '../../components/Common/EmptyContent';

describe('Empty', () => {
  const minProps = {
    children: [
      'abc',
      'def'
    ]
  };
  it('renders correct content', () => {
    const {asFragment, getAllByTestId} = render(<Empty children={minProps.children}/>);
    expect(getAllByTestId('empty')).toHaveLength(2);
    expect(asFragment()).toMatchSnapshot();
  });
});
