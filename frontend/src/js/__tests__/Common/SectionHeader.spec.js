import {render} from '@testing-library/react';
import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import SectionHeader from '../../components/Common/SectionHeader';

describe('Section Header', () => {
  const minProps = {
    breadcrumbs: [
      {
        path: '/apps',
        label: 'Applications'
      }
    ],
    title: 'Flatcar'
  };
  it('renders section header correctly', () => {
    const {asFragment} = render(
      <BrowserRouter>
        <SectionHeader {...minProps}/>
      </BrowserRouter>);
    expect(asFragment()).toMatchSnapshot();
  });
});
