import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import Label from '../../components/common/Label';

describe('Label', () => {
  const minProps = {
    children: 'dummy label',
  };
  it('should render correct label', () => {
    const { asFragment, getByText } = render(<Label children={minProps.children} />);
    expect(asFragment()).toMatchSnapshot();
    expect(getByText(minProps.children)).toBeTruthy();
  });
});
