import { Meta, Story } from '@storybook/react/types-6-0';
import AutoCompletePicker, { AutoCompletePickerProps } from './AutoCompletePicker';

export default {
  title: 'AutoCompletePicker',
  argTypes: {
    onSelect: { action: 'onSelect' },
    onValueChanged: { action: 'onValueChanged' },
  },
} as Meta;

const AutoCompletePickerTemplate: Story<AutoCompletePickerProps> = args => (
  <AutoCompletePicker {...args} />
);

export const Closed = AutoCompletePickerTemplate.bind({});
Closed.args = {
  defaultValue: '2261.0.0',
  suggestions: [
    { primary: '2261.0.0', secondary: 'created: 09/13/2019' },
    { primary: '2247.99.0', secondary: 'created: 09/05/2019' },
    { primary: '2247.2.0', secondary: 'created: 09/13/2019' },
    { primary: '2191.5.0', secondary: 'created: 09/05/2019' },
  ],
  label: 'Package',
  placeholder: 'Pick a package',
  dialogTitle: 'Choose a package',
  pickerPlaceholder: 'Start typing to search a package',
  initialOpen: false,
};

export const Open = AutoCompletePickerTemplate.bind({});
Open.args = {
  ...Closed.args,
  initialOpen: true,
};
