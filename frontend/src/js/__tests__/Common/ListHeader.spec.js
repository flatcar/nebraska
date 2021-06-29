import '../../i18n/config.ts';
import { render } from '@testing-library/react';
import React from 'react';
import ListHeader from '../../components/Common/ListHeader';
import ModalButton from '../../components/Common/ModalButton';

describe('List Header', () => {
  const minProps = {
    title: 'Applications',
    actions: [<ModalButton modalToOpen="AddApplicationModal" data={{ applications: [] }} />],
  };
  it('should render correct List Header title', () => {
    const { getByText } = render(<ListHeader title={minProps.title} />);
    expect(getByText(minProps.title)).toBeInTheDocument();
  });
  it('should render correct List Header actions', () => {
    const { asFragment } = render(<ListHeader actions={minProps.actions} />);
    expect(asFragment()).toMatchSnapshot();
  });
});
