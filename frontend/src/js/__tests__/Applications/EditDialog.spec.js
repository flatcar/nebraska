import {render} from '@testing-library/react';
import React from 'react';
import EditDialog from '../../components/Applications/EditDialog';

describe('Application Edit Dialog', () => {
  it('should render Update button', () => {
    const {getByText} = render(
      <EditDialog
        create = {false}
        data={{}}
        show
      />);
    expect(getByText('Update')).toBeInTheDocument();
    expect(getByText('Update Application')).toBeInTheDocument();
  });
  it('should render Add button', () => {
    const {getByText} = render(
      <EditDialog create
        data={{}}
        show
      />);
    expect(getByText('Add')).toBeInTheDocument();
  });
});
