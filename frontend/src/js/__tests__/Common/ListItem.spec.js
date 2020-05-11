import {render} from '@testing-library/react';
import React from 'react';
import ListItem from '../../components/Common/ListItem';

describe('List Item', () => {
  it('should render correct list item', () => {
    const {getByTestId} = render(<ListItem/>);
    expect(getByTestId('list-item-divider')).toBeInTheDocument();
    expect(getByTestId('list-item')).toBeInTheDocument();
  });
});
