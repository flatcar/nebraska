import { render } from '@testing-library/react';
import React from 'react';
import Loader from '../../components/common/Loader';

describe('Loader', () => {
  it('should render loader without container', () => {
    const { queryByTestId } = render(<Loader noContainer />);
    expect(queryByTestId('loader-container')).toBeFalsy();
  });
  it('should render loader with container', () => {
    const { getByTestId } = render(<Loader />);
    expect(getByTestId('loader-container')).toBeInTheDocument();
  });
});
