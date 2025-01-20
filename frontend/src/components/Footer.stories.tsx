import { createStore } from '@reduxjs/toolkit';
import { Meta, Story } from '@storybook/react';
import { Provider } from 'react-redux';
import { MemoryRouter } from 'react-router-dom';
import FooterComponent from './Footer';

export default {
  title: 'Footer',
  component: FooterComponent,
  argTypes: {},
  decorators: [
    Story => {
      return (
        <MemoryRouter>
          <Story />
        </MemoryRouter>
      );
    },
  ],
} as Meta;

const Template: Story = args => {
  // eslint-disable-next-line no-unused-vars
  const store = createStore((state = { config: {} }, action) => state, {
    config: {
      ...args,
    },
  });
  return (
    <Provider store={store}>
      <FooterComponent />
    </Provider>
  );
};

export const FooterNoOverride = Template.bind({});
FooterNoOverride.args = {
  title: '',
  nebraska_version: '',
};

export const FooterOverride = Template.bind({});
FooterOverride.args = {
  title: 'Some Pro Update Service',
  nebraska_version: '1.2.3',
};
