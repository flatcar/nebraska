import { MuiThemeProvider } from '@material-ui/core/styles';
import {fireEvent, getAllByTestId, render} from '@testing-library/react';
import React from 'react';
import ModalButton from '../../components/Common/ModalButton';
import { theme } from '../../TestHelpers/theme.js';

describe('Modal Button', () => {
  it('should render Application Edit Dialog on Add Icon click', () => {
    const {getByTestId} = render(
      <ModalButton data={{}}
        modalToOpen="AddApplicationModal"
      />);
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('app-edit-form')).toBeInTheDocument();
  });
  it('should render AddGroupModal on Add Icon click', () => {
    const {getByTestId} = render(
      <ModalButton
        data={{}}
        modalToOpen="AddGroupModal"
      />);
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('group-edit-form')).toBeInTheDocument();
  });
  it('should render AddChannelModal on Add Icon click', () => {
    const tree = (
      <MuiThemeProvider theme={theme}>
        <ModalButton data={{}}
          modalToOpen="AddChannelModal"
        />
      </MuiThemeProvider>
    );
    const {getByTestId} = render(tree);
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('channel-edit-form')).toBeInTheDocument();
  });
  it('should render AddPackageModal on Add Icon click', () => {
    const {getByTestId} = render(<ModalButton data={{}} modalToOpen="AddPackageModal"/>);
    fireEvent.click(getByTestId('modal-button'));
    expect(getByTestId('package-edit-form')).toBeInTheDocument();
  });
});
