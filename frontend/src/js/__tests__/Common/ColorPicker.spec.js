import { fireEvent, render } from '@testing-library/react';
import React from 'react';
import { ColorPickerButton } from '../../components/Common/ColorPicker';

describe('Color Picker', () => {
  const minProps = {
    children: 'Click Me',
    color: '#fff',
  };
  it('should render popover on Icon Click', () => {
    const { getByTestId } = render(<ColorPickerButton color={minProps.color} />);
    fireEvent.click(getByTestId('icon-button'));
    expect(getByTestId('popover')).toBeTruthy();
  });
});
