import { Meta, Story } from '@storybook/react/types-6-0';
import React from 'react';
import { CardFeatureLabel, CardFeatureLabelProps } from './Card';

export default {
  title: 'Card',
  component: CardFeatureLabel,
  argTypes: {},
} as Meta;

const Template: Story<CardFeatureLabelProps> = args => <CardFeatureLabel {...args} />;

export const Something = Template.bind({});
Something.args = {
  children: 'meow',
};
