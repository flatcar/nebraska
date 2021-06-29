import '../../i18n/config.ts';
import { render } from '@testing-library/react';
import React from 'react';
import SearchInput from '../../components/Common/ListSearch';

describe('List Search', () => {
  it('should render correct ListSearch', () => {
    const { asFragment } = render(<SearchInput />);
    expect(asFragment()).toMatchSnapshot();
  });
});
